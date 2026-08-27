package projection

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/Roarge/sysml-federation/adapter/model"
)

// snapshotKey is the context key of the model pinned to one operation.
type snapshotKey struct{}

// snapshot is that model. A mutation replaces it with the model it produced,
// so the fields selected on the mutation's result are read from that model.
type snapshot struct{ model *model.Model }

// NewServer builds the executable schema over the store and pins the store's
// current model to every operation, so one query never mixes two versions
// (SR-22). Transports are the caller's: serve adds POST and WebSocket, the
// tests add POST alone.
func NewServer(store *Store) *handler.Server {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{Store: store}}))
	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return next(context.WithValue(ctx, snapshotKey{}, &snapshot{model: store.Current()}))
	})
	return srv
}

// current is the model pinned to the operation, or the store's outside one.
func (r *Resolver) current(ctx context.Context) *model.Model {
	if s, ok := ctx.Value(snapshotKey{}).(*snapshot); ok {
		return s.model
	}
	return r.Store.Current()
}

// advance pins the model a mutation produced.
func advance(ctx context.Context, m *model.Model) {
	if s, ok := ctx.Value(snapshotKey{}).(*snapshot); ok {
		s.model = m
	}
}

// The lookups scan the model's slices rather than calling the model
// package's indexed lookups, so a model built by hand in a test serves
// exactly like a parsed one.

func partByID(m *model.Model, id string) (*model.Part, bool) {
	for _, p := range m.Parts {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}

func requirementByID(m *model.Model, id string) (*model.Requirement, bool) {
	for _, r := range m.Requirements {
		if r.ID == id {
			return r, true
		}
	}
	return nil, false
}

func verificationCaseByID(m *model.Model, id string) (*model.VerificationCase, bool) {
	for _, v := range m.VerificationCases {
		if v.ID == id {
			return v, true
		}
	}
	return nil, false
}

// collect resolves a relationship's ids in order. An id nothing answers to
// is a fault in the model and is reported rather than dropped.
func collect[T any](ids []string, find func(string) (*T, bool), kind string) ([]*T, error) {
	out := make([]*T, 0, len(ids))
	for _, id := range ids {
		v, ok := find(id)
		if !ok {
			return nil, fmt.Errorf("%w: %s %q", model.ErrNotFound, kind, id)
		}
		out = append(out, v)
	}
	return out, nil
}

func parts(m *model.Model, ids []string) ([]*model.Part, error) {
	return collect(ids, func(id string) (*model.Part, bool) { return partByID(m, id) }, "part")
}

func requirements(m *model.Model, ids []string) ([]*model.Requirement, error) {
	return collect(ids, func(id string) (*model.Requirement, bool) { return requirementByID(m, id) }, "requirement")
}

func verificationCases(m *model.Model, ids []string) ([]*model.VerificationCase, error) {
	return collect(ids, func(id string) (*model.VerificationCase, bool) { return verificationCaseByID(m, id) }, "verification case")
}
