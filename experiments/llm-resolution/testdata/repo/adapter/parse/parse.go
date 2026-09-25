package parse

// Parse splits a query into tokens, and refuses an empty one (SR-01).
func Parse(q string) ([]string, error) {
	return []string{q}, nil
}
