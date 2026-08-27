// Package tree is the document's ordered tree of heading, prose and requirement
// nodes, numbered in dotted decimal from the tree alone. It holds requirement
// keys and structure and nothing from the model.
package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Kind is one of the three node kinds.
type Kind string

// The three kinds.
const (
	Heading     Kind = "HEADING"
	Prose       Kind = "PROSE"
	Requirement Kind = "REQUIREMENT"
)

// Node is one entry of the document.
type Node struct {
	ID            string  `json:"id"`
	Kind          Kind    `json:"kind"`
	Text          string  `json:"text,omitempty"`
	RequirementID string  `json:"requirement,omitempty"`
	Children      []*Node `json:"children,omitempty"`
}

// The refusals.
var (
	ErrUnknown     = errors.New("no such node")
	ErrNotExcluded = errors.New("requirement is not excluded")
	ErrCycle       = errors.New("a node cannot be moved under itself")
	ErrProseParent = errors.New("a prose node cannot have children")
	ErrNotText     = errors.New("only a heading or a prose node has text")
)

// Tree is the document. The roots hang off a hidden node so every node has a parent.
type Tree struct {
	root     *Node
	excluded map[string]excluded
	next     int
}

type excluded struct {
	node   *Node
	parent string
}

// Load reads {"nodes": [...]} and checks ids are unique and requirement nodes
// name a requirement. It returns a nil tree with any error, so a caller may
// bind the value before checking the error only once the error is known to be
// nil.
func Load(data []byte) (*Tree, error) {
	var doc struct {
		Nodes []*Node `json:"nodes"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	t := &Tree{root: &Node{Children: doc.Nodes}, excluded: map[string]excluded{}, next: 1}
	seen := map[string]bool{}
	var check func(n *Node) error
	check = func(n *Node) error {
		for _, c := range n.Children {
			if seen[c.ID] {
				return fmt.Errorf("duplicate node id %q", c.ID)
			}
			seen[c.ID] = true
			if c.Kind == Requirement && c.RequirementID == "" {
				return fmt.Errorf("node %q names no requirement", c.ID)
			}
			if c.Kind == Prose && len(c.Children) > 0 {
				return fmt.Errorf("prose node %q has children", c.ID)
			}
			if err := check(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := check(t.root); err != nil {
		return nil, err
	}
	return t, nil
}

// Roots are the top-level nodes in order.
func (t *Tree) Roots() []*Node { return t.root.Children }

// Numbers gives every heading and requirement node its dotted-decimal number.
func (t *Tree) Numbers() map[string]string {
	out := map[string]string{}
	number(t.root, "", out)
	return out
}

func number(n *Node, prefix string, out map[string]string) {
	k := 0
	for _, c := range n.Children {
		if c.Kind == Prose {
			continue
		}
		k++
		out[c.ID] = prefix + strconv.Itoa(k)
		number(c, out[c.ID]+".", out)
	}
}

// Find returns the node with the id, or nil.
func (t *Tree) Find(id string) *Node {
	n, _, _ := t.locate(id)
	return n
}

// Requirement answers for the entity: the number and whether the requirement is in the document.
func (t *Tree) Requirement(requirementID string) (string, bool) {
	n := t.byRequirement(t.root, requirementID)
	if n == nil {
		return "", false
	}
	return t.Numbers()[n.ID], true
}

// Move detaches the node and inserts it under the parent ("" for the root) at
// the index, clamped to the end.
func (t *Tree) Move(id, parentID string, index int) error {
	n, _, _ := t.locate(id)
	if n == nil {
		return fmt.Errorf("%w: %s", ErrUnknown, id)
	}
	parent, err := t.parentFor(parentID)
	if err != nil {
		return err
	}
	if parent == n || contains(n, parent) {
		return ErrCycle
	}
	t.detach(id)
	parent.Children = insert(parent.Children, index, n)
	return nil
}

// InsertHeading puts a new heading where the node is and makes the node its child.
func (t *Tree) InsertHeading(aboveID, text string) (string, error) {
	n, parent, i := t.locate(aboveID)
	if n == nil {
		return "", fmt.Errorf("%w: %s", ErrUnknown, aboveID)
	}
	h := &Node{ID: t.newID(), Kind: Heading, Text: text, Children: []*Node{n}}
	parent.Children[i] = h
	return h.ID, nil
}

// AddProse inserts a prose node under the parent ("" for the root) at the index.
func (t *Tree) AddProse(parentID string, index int, text string) (string, error) {
	parent, err := t.parentFor(parentID)
	if err != nil {
		return "", err
	}
	p := &Node{ID: t.newID(), Kind: Prose, Text: text}
	parent.Children = insert(parent.Children, index, p)
	return p.ID, nil
}

// EditText changes the text of a heading or prose node.
func (t *Tree) EditText(id, text string) error {
	n := t.Find(id)
	if n == nil {
		return fmt.Errorf("%w: %s", ErrUnknown, id)
	}
	if n.Kind == Requirement {
		return ErrNotText
	}
	n.Text = text
	return nil
}

// Exclude removes the requirement's node, promotes its children into its place
// and remembers the former parent for Include.
func (t *Tree) Exclude(requirementID string) error {
	n := t.byRequirement(t.root, requirementID)
	if n == nil {
		return fmt.Errorf("%w: requirement %s", ErrUnknown, requirementID)
	}
	_, parent, i := t.locate(n.ID)
	promoted := append(append([]*Node{}, parent.Children[:i]...), n.Children...)
	parent.Children = append(promoted, parent.Children[i+1:]...)
	n.Children = nil
	t.excluded[requirementID] = excluded{node: n, parent: parent.ID}
	return nil
}

// Include restores an excluded requirement as the last child of its former
// parent, or of the root when that parent is gone.
func (t *Tree) Include(requirementID string) error {
	ex, ok := t.excluded[requirementID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNotExcluded, requirementID)
	}
	parent := t.Find(ex.parent)
	if parent == nil {
		parent = t.root
	}
	parent.Children = append(parent.Children, ex.node)
	delete(t.excluded, requirementID)
	return nil
}

func (t *Tree) parentFor(id string) (*Node, error) {
	if id == "" {
		return t.root, nil
	}
	p := t.Find(id)
	if p == nil {
		return nil, fmt.Errorf("%w: parent %s", ErrUnknown, id)
	}
	if p.Kind == Prose {
		return nil, ErrProseParent
	}
	return p, nil
}

// locate returns the node, its parent and its index among the parent's children.
func (t *Tree) locate(id string) (*Node, *Node, int) {
	var walk func(p *Node) (*Node, *Node, int)
	walk = func(p *Node) (*Node, *Node, int) {
		for i, c := range p.Children {
			if c.ID == id {
				return c, p, i
			}
			if n, parent, j := walk(c); n != nil {
				return n, parent, j
			}
		}
		return nil, nil, -1
	}
	return walk(t.root)
}

func (t *Tree) byRequirement(p *Node, requirementID string) *Node {
	for _, c := range p.Children {
		if c.Kind == Requirement && c.RequirementID == requirementID {
			return c
		}
		if n := t.byRequirement(c, requirementID); n != nil {
			return n
		}
	}
	return nil
}

func (t *Tree) detach(id string) {
	_, parent, i := t.locate(id)
	parent.Children = append(parent.Children[:i:i], parent.Children[i+1:]...)
}

func (t *Tree) newID() string {
	for {
		id := "n" + strconv.Itoa(t.next)
		t.next++
		if t.Find(id) == nil {
			return id
		}
	}
}

func contains(n, target *Node) bool {
	for _, c := range n.Children {
		if c == target || contains(c, target) {
			return true
		}
	}
	return false
}

func insert(list []*Node, index int, n *Node) []*Node {
	index = max(0, min(index, len(list)))
	list = append(list, nil)
	copy(list[index+1:], list[index:])
	list[index] = n
	return list
}
