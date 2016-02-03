package gopherload

import (
	"bufio"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// RequestProfiler will return a RequestProfile after performing an http.Request
type RequestProfiler interface {
	Profile(req *http.Request, tlsConfig *tls.Config) (*RequestProfile, error)
}

// RequestStats represents the data/metrics of an HTTP request
type RequestProfile struct {

	// Start is when the request was started
	Start time.Time

	// DialElapsed is the amount of time it took to connect to the server (from start)
	DialElapsed time.Duration

	// TLSElapsed is the amount of time it took to establish a TLS connection (including connection time)
	// if TLS was not established, this will be equal to DialElapsed.
	TLSElapsed time.Duration

	// SendElapsed is the amount of time spent sending the request data (including connection time)
	SendElapsed time.Duration

	// TTFBElapsed is the amount of time from Start until the first byte was recieved
	TTFBElapsed time.Duration

	// HeadersElapsed is the amount of time until headers were recieved
	HeadersElapsed time.Duration

	// TotalElapsed is the total amount of time of the request
	TotalElapsed time.Duration

	// SentBytes is the total number of bytes sent
	SentBytes int64

	// RecvBytes is the total number of bytes received
	RecvBytes int64

	// RecvBodyBytes is the number of bytes received for headers
	RecvBodyBytes int64

	// StatusCode is the status code received from the server
	StatusCode int
}

// Target is a remote server to perform testing on
type Target string

type writeCounter struct {
	io.Writer
	n int64
}

func (w *writeCounter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.n += int64(n)
	return n, err
}

type readCounter struct {
	io.Reader
	firstByte time.Time
	n         int64
}

func (r *readCounter) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if r.n == 0 && n > 0 {
		r.firstByte = time.Now()
	}
	r.n += int64(n)
	return n, err
}

// Profile will perform an HTTP request against Target, tracking metrics to a RequestProfile struct
// if tlsConfig is not nil, HTTPS will be used instead
func (t Target) Profile(req *http.Request, tlsConfig *tls.Config) (*RequestProfile, error) {
	var stats RequestProfile
	host := string(t)
	if !strings.ContainsRune(host, ':') {
		if tlsConfig == nil {
			host += ":80"
		} else {
			host += ":443"
		}
	}
	stats.Start = time.Now()
	c, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	stats.DialElapsed = time.Since(stats.Start)
	defer c.Close()
	if tlsConfig != nil {
		tcli := tls.Client(c, tlsConfig)
		err = tcli.Handshake()
		if err != nil {
			return nil, err
		}
		stats.TLSElapsed = time.Since(stats.Start)
		defer tcli.Close()
		c = tcli
	} else {
		stats.TLSElapsed = stats.DialElapsed
	}
	w := &writeCounter{Writer: c}
	err = req.Write(w)
	if err != nil {
		return nil, err
	}
	stats.SendElapsed = time.Since(stats.Start)
	rc := &readCounter{Reader: c}
	r := bufio.NewReader(rc)
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, err
	}
	stats.HeadersElapsed = time.Since(stats.Start)
	defer resp.Body.Close()
	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, err
	}
	stats.TotalElapsed = time.Since(stats.Start)
	stats.RecvBodyBytes = n
	stats.RecvBytes = rc.n
	stats.SentBytes = w.n
	stats.TTFBElapsed = rc.firstByte.Sub(stats.Start)
	stats.StatusCode = resp.StatusCode
	return &stats, nil
}
