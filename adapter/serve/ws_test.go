package serve

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/Roarge/sysml-federation/internal/assert"
)

type wsMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// subscribe opens a graphql-transport-ws subscription against a running
// httptest server and returns a function that yields the next event's payload.
func subscribe(t *testing.T, httpURL, query string) func() json.RawMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	// The handshake response body is the library's to close, and on a successful
	// dial it has already been taken off the response, so there is none here.
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(httpURL, "http")+"/graphql", //nolint:bodyclose // closed by the library
		&websocket.DialOptions{Subprotocols: []string{"graphql-transport-ws"}})
	assert.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(websocket.StatusNormalClosure, "") })
	assert.Equal(t, conn.Subprotocol(), "graphql-transport-ws")
	assert.NoError(t, wsjson.Write(ctx, conn, wsMessage{Type: "connection_init"}))
	var ack wsMessage
	assert.NoError(t, wsjson.Read(ctx, conn, &ack))
	assert.Equal(t, ack.Type, "connection_ack")
	raw, err := json.Marshal(struct {
		Query string `json:"query"`
	}{query})
	payload := assert.Must(t, raw, err)
	assert.NoError(t, wsjson.Write(ctx, conn, wsMessage{ID: "1", Type: "subscribe", Payload: payload}))
	return func() json.RawMessage {
		for {
			var m wsMessage
			assert.NoError(t, wsjson.Read(ctx, conn, &m))
			switch m.Type {
			case "next":
				return m.Payload
			case "error", "complete":
				t.Fatalf("unexpected %s: %s", m.Type, m.Payload)
			}
		}
	}
}

func waitForSubscribers(t *testing.T, count func() int, want int) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if count() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("subscribers never reached %d", want)
}
