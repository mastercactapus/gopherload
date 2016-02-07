package gopherload

import "net/http"

type RequestGenerator struct {
	Source  RequestSource
	reqCh   chan *http.Request
	closeCh chan struct{}
}

func (g *RequestGenerator) Start() chan *http.Request {
	g.reqCh = make(chan *http.Request, 1)
	g.closeCh = make(chan struct{})
	go g.loop()
	return g.reqCh
}
func (g *RequestGenerator) Close() error {
	close(g.closeCh)
	return nil
}

func (g *RequestGenerator) loop() {
	var req *http.Request
	var err error
	req, err = g.Source.NewRequest()
	if err != nil {
		panic(err)
	}
	for {
		select {
		case <-g.closeCh:
			return
		default:
		}

		select {
		case <-g.closeCh:
			return
		case g.reqCh <- req:
			req, err = g.Source.NewRequest()
			if err != nil {
				panic(err)
			}
		default:
		}
	}
}
