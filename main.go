package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dimchansky/ipfs-add/config"
	"github.com/dimchansky/ipfs-add/ipfs"
	"github.com/dimchansky/ipfs-add/pathadder"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func run() error {
	cfg := config.Parse()

	c, err := ipfs.New(cfg.IPFSNode)
	if err != nil {
		return err
	}

	a := pathadder.New(c, cfg.HandleHiddenFiles)

	ctx := context.Background()
	for _, p := range cfg.Paths {
		if err := a.AddPath(ctx, p); err != nil {
			return err
		}
	}

	return nil
}
