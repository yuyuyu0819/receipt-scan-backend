package http

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[Request] method=%s path=%s body_read_error=%v", r.Method, r.URL.Path, err)
		} else {
			log.Printf("[Request] method=%s path=%s body=%s", r.Method, r.URL.Path, string(bodyBytes))
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("[Response] method=%s path=%s status=%d body=%s", r.Method, r.URL.Path, status, recorder.body.String())
		if status >= http.StatusBadRequest {
			log.Printf("[Error] method=%s path=%s status=%d error=%s", r.Method, r.URL.Path, status, recorder.body.String())
		}
	})
}
