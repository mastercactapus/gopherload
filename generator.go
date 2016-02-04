package gopherload

import "net/http"

type RequestGenerator struct {
	Source  RequestSource
	reqCh   chan *http.Request
	closeCh chan struct{}
}

func (g *RequestGenerator) Start(n int) chan *http.Request {
	g.reqCh = make(chan *http.Request, n)
	g.closeCh = make(chan struct{})
	for i := 0; i < n; i++ {
		go g.loop()
	}
	return g.reqCh
}
func (g *RequestGenerator) Close() error {
	close(g.closeCh)
	return nil
}

func (g *RequestGenerator) loop() {
	var req *http.Request
	var err error
	for {
		select {
		case <-g.closeCh:
			return
		default:
		}
		req, err = g.Source.NewRequest()
		if err != nil {
			panic(err)
		}
		select {
		case <-g.closeCh:
			return
		case g.reqCh <- req:
		default:
		}
	}
}
