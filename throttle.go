package gopherload

import (
	"net/http"
	"time"
)

type Throttle struct {
	rateCh  chan float64
	reqCh   chan *http.Request
	closeCh chan struct{}
}

func NewThrottle(ch chan *http.Request, rps float64) *Throttle {
	t := &Throttle{
		rateCh:  make(chan float64),
		reqCh:   make(chan *http.Request, 1),
		closeCh: make(chan struct{}),
	}
	go t.loop(ch, rps)
	return t
}

func (t *Throttle) loop(src chan *http.Request, rps float64) {
	delay := time.Duration(float64(time.Second) / rps)
	tc := time.NewTicker(delay / 2)
	last := time.Now()
	bucket := time.Duration(0)
	for {
		select {
		case <-tc.C:
			goto send
		case <-t.closeCh:
			tc.Stop()
			close(t.reqCh)
			return
		case rate := <-t.rateCh:
			delay = time.Duration(float64(time.Second) / rate)
			tc.Stop()
			tc = time.NewTicker(delay / 2)
			if bucket > delay {
				bucket = delay
			}
			goto send
		}
		continue
	send:
		tm := time.Now()
		bucket += tm.Sub(last)
		last = tm
		if bucket >= delay {
			select {
			case t.reqCh <- <-src:
				bucket -= delay
			default:
			}
		}
	}
}

func (t *Throttle) Output() chan *http.Request {
	return t.reqCh
}
func (t *Throttle) Close() error {
	close(t.closeCh)
	close(t.rateCh)
	return nil
}
func (t *Throttle) SetRate(rps float64) {
	t.rateCh <- rps
}
