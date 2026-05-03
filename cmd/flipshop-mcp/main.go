package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gambarini/flip-shop/utils/mcp"
)

func run() error {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	cfg, err := mcp.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	s := mcp.NewServer(logger, cfg)
	return s.Start(context.Background())
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
