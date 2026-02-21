package main

import (
	"context"
	"log"
	"os"

	"github.com/gambarini/flip-shop/utils/mcp"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("flipshop-mcp: starting (scaffold)")

	// Load configuration from environment with defaults and validation
	cfg, err := mcp.LoadFromEnv()
	if err != nil {
		logger.Fatalf("flipshop-mcp: configuration error: %v", err)
	}

	// Create a stub server and start it (no-op for now)
	s := mcp.NewServer(logger, cfg)
	_ = s.Start(context.Background())

	logger.Println("flipshop-mcp: exiting (scaffold)")
}
