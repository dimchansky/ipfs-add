package ipfs_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dimchansky/ipfs-add/ipfs"
	"github.com/dnaeon/go-vcr/recorder"
)

func ExampleIPFS() {
	// start http recorder
	r, err := recorder.New("fixtures/dag-links")
	if err != nil {
		panic(err)
	}
	defer func() { _ = r.Stop() }() // Make sure recorder is stopped once done with it

	c, err := ipfs.NewWithClient("https://ipfs.infura.io:5001", &http.Client{Transport: r})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	file1, err := c.Add(ctx, strings.NewReader("file 1 content"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("file1.txt: %v\n", file1)

	file2, err := c.Add(ctx, strings.NewReader("file 2 content"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("file2.txt: %v\n", file2)

	subDir, err := c.DagPutLinks(ctx, []ipfs.Link{
		file1.ToLink("file1.txt"),
		file2.ToLink("file2.txt"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("subdir: Hash: %v\n", subDir.String())

	subDirStat, err := c.ObjectStat(ctx, subDir.String())
	if err != nil {
		panic(err)
	}
	fmt.Printf("subdir: Size: %v\n", subDirStat.CumulativeSize)

	readmeFile, err := c.Add(ctx, strings.NewReader("README blah blah blah..."))
	if err != nil {
		panic(err)
	}
	fmt.Printf("readme.txt: %v\n", readmeFile)

	rootDir, err := c.DagPutLinks(ctx, []ipfs.Link{
		subDir.ToLink("subdir", subDirStat.CumulativeSize),
		readmeFile.ToLink("README.txt"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("rootdir: Hash: %v\n", rootDir.String())

	rc, err := c.Cat(ctx, rootDir.String()+"/subdir/file1.txt")
	if err != nil {
		panic(err)
	}

	file1Content, err := copyToString(rc)
	if err != nil {
		panic(err)
	}
	fmt.Println(file1Content)

	// Output:
	// file1.txt: Hash: QmSFEbC6Y17cdti7damkjoqESWftkyfSXjdKDQqnf4ECV7 Size: 22
	// file2.txt: Hash: QmVssUfKob8KkUyUiwzoGqNTKqyaEXfqxeGiUJ7ZGyfPLV Size: 22
	// subdir: Hash: QmaRt7pb5LE7991M94XzVCcZgUKPLoihD941GyFSuYBQ9Y
	// subdir: Size: 150
	// readme.txt: Hash: QmeQVyGZQbArEEzukCAXbNbLBBwiBwD4E1WinEeAVw1dZ8 Size: 32
	// rootdir: Hash: QmZgrVVH6Dp4MR4FoF5npDTtJzLneH2KvxoPHixAiucVJH
	// file 1 content
}

func copyToString(r io.Reader) (res string, err error) {
	if c, ok := r.(io.Closer); ok {
		defer func() { _ = c.Close() }()
	}

	var sb strings.Builder
	if _, err = io.Copy(&sb, r); err == nil {
		res = sb.String()
	}
	return
}
