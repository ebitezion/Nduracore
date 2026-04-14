package httpapi

import "net/http"

type StatusResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (srw *StatusResponseWriter) WriteHeader(code int) {
	srw.StatusCode = code
	srw.ResponseWriter.WriteHeader(code)
}
