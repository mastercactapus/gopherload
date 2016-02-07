package gopherload

import (
	"crypto/tls"
	"net/http"
	"runtime"
)

type Result struct {
	Profile *RequestProfile
	Req     *http.Request
	Err     error
}

type LoadGenerator struct {
	Source    chan *http.Request
	Profiler  RequestProfiler
	CLimit    int
	TLSConfig *tls.Config
	resultsCh chan *Result
	closeCh   chan struct{}
	limitCh   chan struct{}
}

func (g *LoadGenerator) loop() {
	for {
		select {
		case <-g.closeCh:
			return
		default:
		}
		req := <-g.Source
		g.limitCh <- struct{}{}
		go g.profile(req)
	}
}

func (g *LoadGenerator) profile(req *http.Request) {
	res := &Result{Req: req}
	res.Profile, res.Err = g.Profiler.Profile(req, g.TLSConfig)
	<-g.limitCh
	g.resultsCh <- res
}

func (g *LoadGenerator) Start(resCh chan *Result) {
	g.resultsCh = resCh
	g.closeCh = make(chan struct{})
	g.limitCh = make(chan struct{}, g.CLimit)
	for i := 0; i < runtime.NumCPU(); i++ {
		go g.loop()
	}
}
func (g *LoadGenerator) Stop() {
	close(g.closeCh)
}
