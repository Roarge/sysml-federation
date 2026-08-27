//go:generate go tool gqlgen generate

// Package document is the example's document subgraph: an editorial tree over
// requirement keys, its numbering, a version counter and its change events.
// It never reads the model.
package document

// Resolver is the root of the resolvers. It holds the service and nothing else.
type Resolver struct{ service *Service }

// deref reads an optional identifier, where the empty string means the root.
// It lives here rather than beside the resolvers that call it, because gqlgen
// moves a helper out of a file it regenerates.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
