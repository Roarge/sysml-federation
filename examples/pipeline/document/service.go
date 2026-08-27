package document

import (
	"context"
	_ "embed"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/coder/websocket"

	"github.com/Roarge/sysml-federation/examples/pipeline/document/tree"
)

//go:embed shipped.json
var shipped []byte

// Service holds the tree, the version and the subscribers.
type Service struct {
	mu      sync.Mutex
	tree    *tree.Tree
	version int
	shipped []byte
	subs    map[chan int]struct{}
}

// New loads the shipped tree.
func New() (*Service, error) { return NewFrom(shipped) }

// NewFrom loads a tree from JSON in the shape of shipped.json.
func NewFrom(data []byte) (*Service, error) {
	t, err := tree.Load(data)
	if err != nil {
		return nil, err
	}
	return &Service{tree: t, version: 1, shipped: data, subs: map[chan int]struct{}{}}, nil
}

// Version is the current document version.
func (s *Service) Version() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

// Subscribers counts the open subscriptions.
func (s *Service) Subscribers() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

// apply runs one operation under the lock, bumps the version and publishes it.
// A slow subscriber keeps only the latest version: the channel holds one
// value and an unread one is dropped before the new one is sent.
func (s *Service) apply(op func(*tree.Tree) error) (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := op(s.tree); err != nil {
		return nil, err
	}
	s.version++
	for ch := range s.subs {
		select {
		case <-ch:
		default:
		}
		ch <- s.version
	}
	return s.snapshot(), nil
}

// reset reloads the shipped tree. The counter keeps growing.
func (s *Service) reset(t *tree.Tree) error {
	fresh, err := tree.Load(s.shipped)
	if err != nil {
		return err
	}
	*t = *fresh
	return nil
}

func (s *Service) document() *Document {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot()
}

// snapshot converts the tree into the served shape. The caller holds the lock.
func (s *Service) snapshot() *Document {
	numbers := s.tree.Numbers()
	var convert func(nodes []*tree.Node) []*Node
	convert = func(nodes []*tree.Node) []*Node {
		out := make([]*Node, 0, len(nodes))
		for _, n := range nodes {
			node := &Node{ID: n.ID, Kind: NodeKind(n.Kind), Children: convert(n.Children)}
			if n.Kind != tree.Requirement {
				text := n.Text
				node.Text = &text
			}
			if num, ok := numbers[n.ID]; ok {
				node.Number = &num
			}
			if n.Kind == tree.Requirement {
				node.Requirement = &Requirement{ID: n.RequirementID, DocumentNumber: node.Number, Included: true}
			}
			out = append(out, node)
		}
		return out
	}
	return &Document{Version: s.version, Nodes: convert(s.tree.Roots())}
}

func (s *Service) requirement(id string) *Requirement {
	s.mu.Lock()
	defer s.mu.Unlock()
	number, included := s.tree.Requirement(id)
	r := &Requirement{ID: id, Included: included}
	if included {
		r.DocumentNumber = &number
	}
	return r
}

// subscribe registers a channel that receives every new version until ctx ends.
func (s *Service) subscribe(ctx context.Context) <-chan int {
	ch := make(chan int, 1)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subs, ch)
		close(ch)
		s.mu.Unlock()
	}()
	return ch
}

// Handler serves the subgraph: /graphql over POST and WebSocket, and /health.
func Handler(s *Service) http.Handler {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{service: s}}))
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Implementation:        transport.CoderWebsocketImplementation{AcceptOptions: websocket.AcceptOptions{InsecureSkipVerify: true}},
	})
	srv.Use(extension.Introspection{})
	mux := http.NewServeMux()
	mux.Handle("/graphql", srv)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}
