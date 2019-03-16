package pathadder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dimchansky/ipfs-add/ipfs"
)

// PathAdder adds path to IPFS
type PathAdder struct {
	ipfsClient        *ipfs.IPFS
	handleHiddenFiles bool
}

// New creates new instance of PathAdder
func New(ipfsClient *ipfs.IPFS, handleHiddenFiles bool) *PathAdder {
	return &PathAdder{
		ipfsClient:        ipfsClient,
		handleHiddenFiles: handleHiddenFiles,
	}
}

// AddPath adds path to IPFS. Directories are added recursively.
func (a *PathAdder) AddPath(ctx context.Context, fPath string) error {
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

	_, err = a.add(ctx, stat, path.Base(fPath), fPath)
	return err
}

func (a *PathAdder) add(ctx context.Context, fi os.FileInfo, fName, fPath string) (*ipfs.AddResult, error) {
	if fi.IsDir() {
		return a.addDir(ctx, fName, fPath)
	}
	return a.addFile(ctx, fName, fPath)
}

func (a *PathAdder) addDir(ctx context.Context, fName, fPath string) (*ipfs.AddResult, error) {
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

		fileName := filepath.ToSlash(filepath.Join(fName, shortFName))
		filePath := filepath.ToSlash(filepath.Join(fPath, shortFName))

		res, err := a.add(ctx, f, fileName, filePath)
		if err != nil {
			return nil, err
		}

		links = append(links, res.ToLink(shortFName))
	}

	dir, err := a.ipfsClient.DagPutLinks(ctx, links)
	if err != nil {
		return nil, err
	}

	dirStat, err := a.ipfsClient.ObjectStat(ctx, dir.String())
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

func (a *PathAdder) addFile(ctx context.Context, fName, fPath string) (*ipfs.AddResult, error) {
	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	res, err := a.ipfsClient.Add(ctx, file)
	if err != nil {
		return nil, err
	}

	a.addedEvent(fName, res)
	return res, nil
}

func (a *PathAdder) addedEvent(name string, res *ipfs.AddResult) {
	fmt.Printf("added %v %v\n", res.Hash, name)
}
