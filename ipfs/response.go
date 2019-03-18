package ipfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

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
