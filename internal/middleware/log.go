package middleware

import (
	"net/http"
	"time"

	"go.mattglei.ch/timber"
)

// wrappedWriter provides a custom interface that allows us to store the status code of a request
// when it is being handled by our mux
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}

func (w *wrappedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)
		timber.InfoSince(
			start,
			"handled request",
			timber.A("code", wrapped.statusCode),
			timber.A("path", r.URL.Path),
		)
	})
}
