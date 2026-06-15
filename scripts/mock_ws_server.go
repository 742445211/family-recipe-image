// Mock WSS server for local gateway testing.
// Usage: go run ./scripts/mock_ws_server.go
package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func main() {
	token := os.Getenv("GATEWAY_TOKEN")
	if token == "" {
		token = "test-token"
	}
	addr := ":8443"
	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")

	http.HandleFunc("/api/ws/image-worker", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteJSON(map[string]string{"type": "registered", "worker_id": r.URL.Query().Get("worker_id")})
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var msg map[string]any
			_ = json.Unmarshal(data, &msg)
			log.Printf("recv: %s", string(data))
			if msg["type"] == "ping" {
				_ = conn.WriteJSON(map[string]string{"type": "pong"})
			}
		}
	})

	if cert != "" && key != "" {
	 srv := &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}
		log.Printf("mock WSS on %s", addr)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	}
	log.Printf("mock WS (no TLS) on %s - use TLS_CERT/TLS_KEY for WSS", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
