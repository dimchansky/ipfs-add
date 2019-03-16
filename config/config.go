package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	// Version (set by compiler) is the version of program
	Version = "undefined"
	// BuildTime (set by compiler) is the program build time in '+%Y-%m-%dT%H:%M:%SZ' format
	BuildTime = "undefined"
	// GitHash (set by compiler) is the git commit hash of source tree
	GitHash = "undefined"
)

// Config holds parsed command line parameters
type Config struct {
	Paths             []string
	IPFSNode          string
	HandleHiddenFiles bool
}

// Parse parses the command-line flags from os.Args[1:] and returns parsed configuration.
func Parse() *Config {
	cfg := &Config{
		IPFSNode:          "https://ipfs.infura.io:5001",
		HandleHiddenFiles: false,
	}
	printVersion := false

	flag.StringVar(&cfg.IPFSNode, "node", cfg.IPFSNode, "The url of IPFS node to use.")
	flag.BoolVar(&cfg.HandleHiddenFiles, "H", cfg.HandleHiddenFiles, "Include files that are hidden. Only takes effect on directory add.")
	flag.BoolVar(&printVersion, "v", printVersion, "Print program version.")

	flag.CommandLine.Usage = func() {
		out := flag.CommandLine.Output()
		_, _ = fmt.Fprintf(out, `USAGE:
  %s [options] <path>...

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

	if printVersion {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\tBuildTime: %v\tGitHash: %s\n", Version, BuildTime, GitHash)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "the <path> to a file is not provided")
		flag.CommandLine.Usage()
		os.Exit(3)
	}

	cfg.Paths = flag.Args()

	return cfg
}
