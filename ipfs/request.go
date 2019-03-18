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
	mpf, err := toMultipartFile(r.body)
	if err != nil {
		return
	}

	var body io.Reader
	if mpf != nil {
		body = mpf.Body
	}
	req, err := http.NewRequest("POST", r.getURL(), body)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	if mpf != nil {
		req.Header.Set("Content-Type", mpf.ContentType)
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

type multiPartFile struct {
	ContentType string
	Body        io.Reader
}

func toMultipartFile(f io.Reader) (mpf *multiPartFile, err error) {
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

	mpf = &multiPartFile{
		ContentType: w.FormDataContentType(),
		Body:        body,
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
