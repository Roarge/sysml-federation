//go:generate go tool gqlgen generate

// Package projection is the adapter's GraphQL subgraph: the schema in
// ../schema.graphql bound to the adapter/model types, a store that holds
// the current model and its version, and the resolvers for relationships,
// mutations and the version event. It holds nothing the schema does not
// show.
package projection

// Resolver is the root of the generated resolvers.
type Resolver struct{ Store *Store }
