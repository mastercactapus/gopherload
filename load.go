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
	Source    RequestSource
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
		req, err := g.Source.NewRequest()
		if err != nil {
			panic(err)
		}
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

func (g *LoadGenerator) Start(buffer int) chan *Result {
	g.resultsCh = make(chan *Result, buffer)
	g.closeCh = make(chan struct{})
	g.limitCh = make(chan struct{}, g.CLimit)
	for i := 0; i < runtime.NumCPU(); i++ {
		go g.loop()
	}
	return g.resultsCh
}
func (g *LoadGenerator) Stop() {
	close(g.closeCh)
}
