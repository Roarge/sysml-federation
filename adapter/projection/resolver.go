//go:generate go tool gqlgen generate

// Package projection is the adapter's GraphQL subgraph: the schema in
// ../schema.graphql bound to the adapter/model types, a store that holds
// the current model and its version, and the resolvers for relationships,
// mutations and the version event. It holds nothing the schema does not
// show.
package projection

// Resolver is the root of the generated resolvers.
type Resolver struct{ Store *Store }

// optional turns a model string the schema types as nullable into the value the
// graph publishes. The model has one spelling for an absent word, the empty
// string, and a client that reads an empty string back cannot tell that the
// source said nothing. Absent is published as null instead.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
