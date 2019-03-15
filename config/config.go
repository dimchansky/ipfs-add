package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Paths             []string
	Node              string
	HandleHiddenFiles bool
}

func Parse() *Config {
	cfg := &Config{
		Node:              "https://ipfs.infura.io:5001",
		HandleHiddenFiles: false,
	}

	flag.StringVar(&cfg.Node, "node", cfg.Node, "The url of IPFS node to use.")
	flag.BoolVar(&cfg.HandleHiddenFiles, "H", cfg.HandleHiddenFiles, "Include files that are hidden. Only takes effect on directory add.")

	flag.CommandLine.Usage = func() {
		out := flag.CommandLine.Output()
		_, _ = fmt.Fprintf(out, `USAGE:
  %s: [options] <path>...

ARGUMENTS

  <path>... - The path to a file to be added to ipfs.

OPTIONS

`, os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(out, `
DESCRIPTION

  Adds contents of <path> to ipfs. Note that directories are added recursively, to form the ipfs
  MerkleDAG.
`)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "the <path> to a file is not provided")
		flag.CommandLine.Usage()
		os.Exit(3)
	}

	cfg.Paths = flag.Args()

	return cfg
}
