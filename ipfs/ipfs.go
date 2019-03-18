package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// IPFS provides limited functionality to interact with the IPFS (https://ipfs.io)
type IPFS struct {
	url     *url.URL
	httpcli *http.Client
}

// New creates new instance of IPFS from the provided url
func New(url string) (*IPFS, error) {
	c := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	}

	return NewWithClient(url, c)
}

// NewWithClient creates new instance of IPFS from the provided url and http.Client
func NewWithClient(uri string, c *http.Client) (*IPFS, error) {
	if !strings.HasPrefix(uri, "http") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	return &IPFS{
		url:     u,
		httpcli: c,
	}, nil
}

// AddResult contains result of AddResult command
type AddResult struct {
	Hash string `json:"Hash"`
	Size uint64 `json:"Size,string"`
}

// Link represents an IPFS Merkle DAG Link between Nodes.
type Link struct {
	// multihash of the target object
	Cid Cid `json:"Cid"`
	// utf string name. should be unique per object
	Name string `json:"Name"`
	// cumulative size of target object
	Size uint64 `json:"Size"`
}

// String implements fmt.Stringer interface
func (a *AddResult) String() string { return fmt.Sprintf("Hash: %s Size: %d", a.Hash, a.Size) }

// ToLink creates link from AddResult
func (a *AddResult) ToLink(name string) Link { return Link{Cid: Cid(a.Hash), Name: name, Size: a.Size} }

// Add a file to ipfs from the given reader, returns the hash of the added file
func (f *IPFS) Add(ctx context.Context, r io.Reader) (*AddResult, error) {
	var out AddResult
	return &out, f.request("add").
		Option("progress", false).
		Option("pin", true).
		Body(r).
		Exec(ctx, &out)
}

// Cat the content at the given path. Callers need to drain and close the returned reader after usage.
func (f *IPFS) Cat(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := f.request("cat", path).Send(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

type dagNode struct {
	Data  string `json:"data"`
	Links []Link `json:"links"`
}

// DagPutLinks puts directory containing links and returns Cid of directory.
func (f *IPFS) DagPutLinks(ctx context.Context, links []Link) (Cid, error) {
	dagJSONBytes, err := json.Marshal(dagNode{
		Data:  "CAE=",
		Links: links,
	})
	if err != nil {
		return "", err
	}

	var out struct {
		Cid Cid `json:"Cid"`
	}

	return out.Cid, f.request("dag/put").
		Option("format", "protobuf").
		Option("input-enc", "json").
		Option("pin", true).
		Body(bytes.NewBuffer(dagJSONBytes)).
		Exec(ctx, &out)
}

// DagGetLinks gets directory links.
func (f *IPFS) DagGetLinks(ctx context.Context, cid Cid) ([]Link, error) {
	var out dagNode
	return out.Links, f.request("dag/get", cid.String()).Exec(ctx, &out)
}

// ObjectStat provides information about dag nodes
type ObjectStat struct {
	Hash           string `json:"Hash"`
	NumLinks       uint64 `json:"NumLinks"`
	BlockSize      uint64 `json:"BlockSize"`
	LinksSize      uint64 `json:"LinksSize"`
	DataSize       uint64 `json:"DataSize"`
	CumulativeSize uint64 `json:"CumulativeSize"`
}

// ObjectStat returns information about the dag node
func (f *IPFS) ObjectStat(ctx context.Context, path string) (*ObjectStat, error) {
	var out ObjectStat
	return &out, f.request("object/stat", path).Exec(ctx, &out)
}
