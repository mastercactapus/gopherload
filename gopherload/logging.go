package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type statusCodeRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusCodeRecorder) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.statusCode = statusCode
}

func requestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		defer func() {
			err := recover()
			if err != nil {
				log.WithFields(log.Fields{
					"ResponseTime": time.Since(start),
					"Method":       req.Method,
					"Path":         req.URL.Path,
				}).Errorln("panic:", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		wr := &statusCodeRecorder{ResponseWriter: w, statusCode: 200}
		h.ServeHTTP(wr, req)
		l := log.WithFields(log.Fields{
			"ResponseTime": time.Since(start),
			"StatusCode":   wr.statusCode,
			"Method":       req.Method,
			"Path":         req.URL.Path,
		})

		if wr.statusCode == 500 {
			l.Warnln("response")
		} else {
			l.Infoln("response")
		}
	})
}
