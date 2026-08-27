//go:generate go tool gqlgen generate

// Package capacity is the example's analysis subgraph. It computes a rollup,
// a bottleneck and verdicts on every read from the fields the router carries
// in its @requires, holds nothing between requests, and knows two words of
// the model: the quantity it computes and the attribute it reads.
package capacity

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"
)

// Resolver holds the two configured names and nothing else.
type Resolver struct{ Names flow.Names }

// Handler serves the subgraph: POST /graphql and GET /health.
func Handler(names flow.Names) http.Handler {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{Names: names}}))
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	mux := http.NewServeMux()
	mux.Handle("/graphql", srv)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}

// subject turns a Part representation into what the flow package reads.
func (r *Resolver) subject(p *Part) flow.Subject {
	s := flow.Subject{Name: p.Name}
	for _, a := range p.Attributes {
		if a.Name == r.Names.Attribute {
			s.HasAttribute, s.Attribute = true, a.Value
		}
	}
	for _, c := range p.Parts {
		s.Children = append(s.Children, flow.Node{ID: c.ID, Name: c.Name, Value: r.attribute(c)})
	}
	for _, c := range p.Connections {
		s.Edges = append(s.Edges, flow.Edge{From: c.From, To: c.To})
	}
	return s
}

func (r *Resolver) attribute(p *Part) *float64 {
	for _, a := range p.Attributes {
		if a.Name == r.Names.Attribute {
			return a.Value
		}
	}
	return nil
}

// verdict is shared by Requirement.verdict and Requirement.verdictReason.
func (r *Resolver) verdict(req *Requirement) (string, string) {
	var quantity, comparison, vc string
	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	if req.Comparison != nil {
		comparison = string(*req.Comparison)
	}
	if len(req.VerifiedBy) > 0 && req.VerifiedBy[0].ShortName != nil {
		vc = *req.VerifiedBy[0].ShortName
	}
	var subject flow.Subject
	if req.Subject != nil {
		subject = r.subject(req.Subject)
	}
	return flow.Verdict(r.Names, quantity, comparison, req.Limit, subject, vc)
}
