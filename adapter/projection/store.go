package projection

import (
	"sync"

	"github.com/Roarge/sysml-federation/adapter/model"
)

// Store holds the current model, the shipped one for reset, the version
// counter and the subscribers. Models are immutable, so a reader keeps the
// pointer it took and a writer swaps in the next one (C-93).
type Store struct {
	mu      sync.Mutex
	shipped *model.Model
	current *model.Model
	version int
	subs    map[chan int]struct{}
}

// NewStore starts from a model whose Version is the first value of the
// counter, 1 for a parsed model.
func NewStore(m *model.Model) *Store {
	return &Store{shipped: m, current: m, version: m.Version, subs: map[chan int]struct{}{}}
}

// Load reads and parses a model file into a new store.
func Load(path string) (*Store, error) {
	m, err := model.Load(path)
	if err != nil {
		return nil, err
	}
	return NewStore(m), nil
}

// Current is the model every read of this instant sees.
func (s *Store) Current() *model.Model {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// Version is the current counter value (C-92).
func (s *Store) Version() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

// SetAttribute patches a part's literal attribute and returns the part of
// the new model. Refusals are the model package's (SR-24, SR-25).
func (s *Store) SetAttribute(partID, name string, value float64) (*model.Part, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := s.current.SetAttribute(partID, name, value)
	if err != nil {
		return nil, err
	}
	s.install(next)
	p, _ := partByID(next, partID)
	return p, nil
}

// SetLimit patches a requirement's literal limit and returns the requirement
// of the new model.
func (s *Store) SetLimit(requirementID string, value float64) (*model.Requirement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := s.current.SetLimit(requirementID, value)
	if err != nil {
		return nil, err
	}
	s.install(next)
	r, _ := requirementByID(next, requirementID)
	return r, nil
}

// Reset serves the shipped model again under the next version. The copy is
// shallow: the parsed structure is never modified, so sharing it is safe,
// and only the version differs (SR-44, C-92).
func (s *Store) Reset() *model.Model {
	s.mu.Lock()
	defer s.mu.Unlock()
	fresh := *s.shipped
	s.install(&fresh)
	return &fresh
}

// install numbers a new model, makes it current and tells the subscribers.
// The caller holds the lock and nobody else holds the pointer yet.
func (s *Store) install(next *model.Model) {
	s.version++
	next.Version = s.version
	s.current = next
	s.notify(s.version)
}
