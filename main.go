package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dimchansky/ipfs-add/config"
	"github.com/dimchansky/ipfs-add/ipfs"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func run() error {
	cfg := config.Parse()

	c := ipfs.New(cfg.Node)
	a := pathAdder{c, false}

	ctx := context.Background()
	for _, p := range cfg.Paths {
		if err := a.AddPath(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

type pathAdder struct {
	c                 *ipfs.IPFS
	handleHiddenFiles bool
}

func (a *pathAdder) AddPath(ctx context.Context, fPath string) error {
	fPath = filepath.ToSlash(filepath.Clean(fPath))
	if fPath == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cwd, err = filepath.EvalSymlinks(cwd)
		if err != nil {
			return err
		}
		fPath = filepath.ToSlash(cwd)
	}

	stat, err := os.Lstat(fPath)
	if err != nil {
		return err
	}

	fName := path.Base(fPath)
	if stat.IsDir() {
		_, err := a.addDir(ctx, fName, fPath)
		return err
	} else {
		_, err := a.addFile(ctx, fName, fPath)
		return err
	}
}

func (a *pathAdder) addDir(ctx context.Context, fName, fPath string) (*ipfs.AddResult, error) {
	files, err := ioutil.ReadDir(fPath)
	if err != nil {
		return nil, err
	}

	links := make([]ipfs.Link, 0, len(files))
	for _, f := range files {
		shortFName := f.Name()
		if !a.handleHiddenFiles && strings.HasPrefix(shortFName, ".") {
			continue
		}

		var addFun func(context.Context, string, string) (*ipfs.AddResult, error)
		if f.IsDir() {
			addFun = a.addDir
		} else {
			addFun = a.addFile
		}

		fileName := filepath.ToSlash(filepath.Join(fName, shortFName))
		filePath := filepath.ToSlash(filepath.Join(fPath, shortFName))
		res, err := addFun(ctx, fileName, filePath)
		if err != nil {
			return nil, err
		}

		links = append(links, res.ToLink(shortFName))
	}

	dir, err := a.c.DagPutLinks(ctx, links)
	if err != nil {
		return nil, err
	}

	dirStat, err := a.c.ObjectStat(ctx, dir.String())
	if err != nil {
		return nil, err
	}

	res := &ipfs.AddResult{
		Hash: dir.String(),
		Size: dirStat.CumulativeSize,
	}

	a.addedEvent(fName, res)
	return res, nil
}

func (a *pathAdder) addFile(ctx context.Context, fName, fPath string) (*ipfs.AddResult, error) {
	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	res, err := a.c.Add(ctx, file)
	if err != nil {
		return nil, err
	}

	a.addedEvent(fName, res)
	return res, nil
}

func (a *pathAdder) addedEvent(name string, res *ipfs.AddResult) {
	fmt.Printf("added %v %v\n", res.Hash, name)
}
