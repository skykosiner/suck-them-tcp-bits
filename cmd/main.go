package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex // Ensure thread-safe access to the clients map
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// Add the new WebSocket connection to the clients map
	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	// Create a cancellable context for this WebSocket connection
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		// Clean up when the WebSocket connection closes
		cancel()
		clientsMu.Lock()
		delete(clients, ws)
		clientsMu.Unlock()
		ws.Close()
	}()

	go func() {
		// This goroutine will listen for connection close or context cancellation
		<-ctx.Done()
		log.Println("WebSocket connection closed")
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			cancel() // Cancel the context, signaling that this connection is done
			break
		}

		log.Printf("Received message: %v", string(msg))

		// Broadcast the message to all clients except the sender
		clientsMu.Lock()
		for client := range clients {
			if client != ws {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("Error writing to client: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
		clientsMu.Unlock()
	}
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "localhost:42069", "the address for the WebSocket server")
	flag.Parse()

	http.HandleFunc("/ws", handleWebSocket)

	server := &http.Server{
		Addr: addr,
	}

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting WebSocket server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	time.Sleep(5 * time.Second)
	fmt.Println("she suck on my socket")

	<-quit // Wait for a termination signal
	for client := range clients {
		client.WriteMessage(websocket.CloseMessage, []byte("It's Joever"))
	}
	log.Println("Shutting down server...")

	// Create a deadline for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully")
}
