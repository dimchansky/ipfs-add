package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type IPFS struct {
	url     string
	httpcli *http.Client
}

// New creates new instance of IPFS from the provided url
func New(url string) *IPFS {
	c := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	}

	return NewWithClient(url, c)
}

// NewWithClient creates new instance of IPFS from the provided url and http.Client
func NewWithClient(url string, c *http.Client) *IPFS {
	return &IPFS{
		url:     url,
		httpcli: c,
	}
}

// AddResult contains result of AddResult command
type AddResult struct {
	Hash string `json:"Hash"`
	Size uint64 `json:"Size,string"`
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

type request struct {
	c       *http.Client
	apiBase string
	command string
	args    []string
	opts    map[string]string
	body    io.Reader
}

func (f *IPFS) request(command string, args ...string) *request {
	uri := f.url
	if !strings.HasPrefix(uri, "http") {
		uri = "http://" + uri
	}

	opts := map[string]string{
		"encoding":        "json",
		"stream-channels": "true",
	}

	return &request{
		c:       f.httpcli,
		apiBase: uri + "/api/v0",
		command: command,
		args:    args,
		opts:    opts,
	}
}

// Option sets the given option.
func (r *request) Option(key string, value interface{}) *request {
	var s string
	switch v := value.(type) {
	case bool:
		s = strconv.FormatBool(v)
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		// slow case.
		s = fmt.Sprint(value)
	}
	if r.opts == nil {
		r.opts = make(map[string]string, 1)
	}
	r.opts[key] = s
	return r
}

// Body sets the request body to the given reader.
func (r *request) Body(body io.Reader) *request {
	r.body = body
	return r
}

func (r *request) Send(ctx context.Context) (resp *response, err error) {
	reqBody, reqContentType, err := toMultipartFile(r.body)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", r.getURL(), reqBody)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	if reqBody != nil {
		req.Header.Set("Content-Type", reqContentType)
	}

	res, err := r.c.Do(req)
	if err != nil {
		return
	}

	contentType := res.Header.Get("Content-Type")
	parts := strings.Split(contentType, ";")
	contentType = parts[0]

	resp = &response{}
	if res.StatusCode >= http.StatusBadRequest {
		e := &responseError{
			Command: r.command,
		}
		resp.Error = e

		switch {
		case res.StatusCode == http.StatusNotFound:
			e.Message = "command not found"
		case contentType == "text/plain":
			out, err := ioutil.ReadAll(res.Body)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "ipfs: warning! response (%d) read error: %s\n", res.StatusCode, err)
			}
			e.Message = string(out)
		case contentType == "application/json":
			if err = json.NewDecoder(res.Body).Decode(e); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "ipfs: warning! response (%d) unmarshall error: %s\n", res.StatusCode, err)
			}
		default:
			_, _ = fmt.Fprintf(os.Stderr, "ipfs: warning! unhandled response (%d) encoding: %s", res.StatusCode, contentType)
			out, err := ioutil.ReadAll(res.Body)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "ipfs: response (%d) read error: %s\n", res.StatusCode, err)
			}
			e.Message = fmt.Sprintf("unknown ipfs error encoding: %q - %q", contentType, out)
		}

		// drain body and close
		_, _ = io.Copy(ioutil.Discard, res.Body)
		_ = res.Body.Close()
	} else {
		resp.Output = res.Body
	}

	return
}

// Exec sends the request a request and decodes the response.
func (r *request) Exec(ctx context.Context, res interface{}) error {
	httpRes, err := r.Send(ctx)
	if err != nil {
		return err
	}

	if res == nil {
		_ = httpRes.Close()
		if httpRes.Error != nil {
			return httpRes.Error
		}
		return nil
	}

	return httpRes.Decode(res)
}

type response struct {
	Output io.ReadCloser
	Error  *responseError
}

func (r *response) Close() error {
	readCloser := r.Output
	if readCloser != nil {
		// always drain output (response body)
		_, _ = io.Copy(ioutil.Discard, readCloser)
		return readCloser.Close()
	}
	return nil
}

func (r *response) Decode(dec interface{}) error {
	defer func() { _ = r.Close() }()
	if r.Error != nil {
		return r.Error
	}

	return json.NewDecoder(r.Output).Decode(dec)
}

type responseError struct {
	Command string
	Message string
	Code    int
}

func (e *responseError) Error() string {
	var out string
	if e.Command != "" {
		out = e.Command + ": "
	}
	if e.Code != 0 {
		out = fmt.Sprintf("%s%d: ", out, e.Code)
	}
	return out + e.Message
}

func toMultipartFile(f io.Reader) (newBody io.Reader, contentType string, err error) {
	if f == nil {
		return
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", "")
	if err != nil {
		return
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return
	}

	err = w.Close()
	if err != nil {
		return
	}

	newBody = body
	contentType = w.FormDataContentType()
	return
}

func (r *request) getURL() string {

	values := make(url.Values)
	for _, arg := range r.args {
		values.Add("arg", arg)
	}
	for k, v := range r.opts {
		values.Add(k, v)
	}

	return fmt.Sprintf("%s/%s?%s", r.apiBase, r.command, values.Encode())
}
