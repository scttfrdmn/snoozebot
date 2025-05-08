package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/scottfridman/snoozebot/agent/api"
	"github.com/scottfridman/snoozebot/agent/store"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	pluginsDir := flag.String("plugins-dir", "/etc/snoozebot/plugins", "Directory containing plugins")
	flag.Parse()

	fmt.Println("Starting Snoozebot Agent")
	fmt.Printf("Listening on port: %d\n", *port)
	fmt.Printf("Plugins directory: %s\n", *pluginsDir)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create store for managing instance state
	instanceStore := store.NewMemoryStore()

	// Create API server
	apiServer := api.NewServer(instanceStore, *pluginsDir)

	// Start API server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		fmt.Printf("API server listening on %s\n", addr)
		
		if err := http.ListenAndServe(addr, apiServer.Router()); err != nil {
			fmt.Printf("API server error: %v\n", err)
			cancel() // Cancel context on server error
		}
	}()

	// Wait for interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case sig := <-sigs:
		fmt.Printf("Received signal: %v\n", sig)
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	}

	fmt.Println("Shutting down...")
}