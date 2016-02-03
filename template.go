package gopherload

import (
	"bytes"
	"io"
	"net/http"
)

// SimpleTemplate is good enough for most request generation
type SimpleTemplate struct {
	Headers http.Header
	URL     string
	Method  string
	Body    []byte
}

// BuildRequest will create a new request from a SimpleTemplate
func (t SimpleTemplate) NewRequest() (*http.Request, error) {
	var b io.Reader
	if t.Body != nil {
		b = bytes.NewReader(t.Body)
	}

	req, err := http.NewRequest(t.Method, t.URL, b)
	if err != nil {
		if bc, ok := b.(io.Closer); ok {
			bc.Close()
		}
		return nil, err
	}
	if t.Headers != nil {
		for hname, vals := range t.Headers {
			v := make([]string, len(vals))
			copy(v, vals)
			req.Header[hname] = v
		}
	}
	return req, nil
}
