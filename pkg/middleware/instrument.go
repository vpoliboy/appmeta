/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package middleware

import (
	"fmt"
	kitexpvar "github.com/go-kit/kit/metrics/expvar"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type interceptingHttpWriter struct {
	delegate   http.ResponseWriter
	statusCode int
}

func (i *interceptingHttpWriter) Header() http.Header {
	return i.delegate.Header()
}

func (i *interceptingHttpWriter) Write(b []byte) (int, error) {
	return i.delegate.Write(b)
}

func (i *interceptingHttpWriter) WriteHeader(statusCode int) {
	i.statusCode = statusCode
	i.delegate.WriteHeader(statusCode)
}

// InstrumentingMiddleware uses the expvar package to instrument the http calls. Currently http latencies
// and statuscode counters are supported.
func InstrumentingMiddleware(label string) mux.MiddlewareFunc {

	httpTotal := kitexpvar.NewCounter(fmt.Sprintf("%s.http.count.total", label))
	http2xx := kitexpvar.NewCounter(fmt.Sprintf("%s.http.count.2xx", label))
	http3xx := kitexpvar.NewCounter(fmt.Sprintf("%s.http.count.3xx", label))
	http4xx := kitexpvar.NewCounter(fmt.Sprintf("%s.http.count.4xx", label))
	http5xx := kitexpvar.NewCounter(fmt.Sprintf("%s.http.count.5xx", label))
	httpLatency := kitexpvar.NewHistogram(fmt.Sprintf("%s.http.latency", label), 50)

	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			httpTotal.Add(1)
			iw := &interceptingHttpWriter{delegate: w}
			defer func(begin time.Time) {
				switch sc := iw.statusCode; {
				case sc >= 200 && sc < 300:
					http2xx.Add(1)
				case sc >= 300 && sc < 400:
					http3xx.Add(1)
				case sc >= 400 && sc < 500:
					http4xx.Add(1)
				case sc >= 500 && sc < 600:
					http5xx.Add(1)
				}
				httpLatency.Observe(time.Since(begin).Seconds() * 1000)

			}(time.Now())

			next.ServeHTTP(iw, r)
		})
	})
}
