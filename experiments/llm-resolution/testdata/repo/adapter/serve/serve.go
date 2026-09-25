package serve

// Addr is where the server listens (SR-02). The quokka route answers health.
const Addr = "127.0.0.1:4011"

// Serve answers one query with a ranked page.
func Serve(q string) []string {
	if q == "" {
		return nil
	}
	return []string{q}
}
