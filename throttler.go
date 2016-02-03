package gopherload

import (
	"net/http"
	"time"
)

type newRequest struct {
	req *http.Request
	err error
}

type ThrottlingRequestSource struct {
	Source  RequestSource
	RPS     float64
	reqCh   chan *http.Request
	closeCh chan struct{}
}

func NewThrottlingRequestSource(source RequestSource, rps float64) *ThrottlingRequestSource {
	s := &ThrottlingRequestSource{
		Source:  source,
		RPS:     rps,
		reqCh:   make(chan *http.Request, 1),
		closeCh: make(chan struct{}),
	}
	go s.loop()
	return s
}

func (s *ThrottlingRequestSource) NewRequest() (*http.Request, error) {
	return <-s.reqCh, nil
}

func (s *ThrottlingRequestSource) loop() {
	t := time.NewTicker(time.Duration(float64(time.Second) / s.RPS))
	var req *http.Request
	var err error
	for {
		select {
		case <-t.C:
			req, err = s.Source.NewRequest()
			if err != nil {
				panic(err)
			}
			select {
			case s.reqCh <- req:
			default:
				// already pending
			}
		case <-s.closeCh:
			t.Stop()
			close(s.reqCh)
			return
		}
	}
}
func (s *ThrottlingRequestSource) Close() error {
	close(s.closeCh)
	return nil
}
