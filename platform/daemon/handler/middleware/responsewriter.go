package middleware

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter

	Status int
	Size   int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,

		Status: 200,
		Size:   0,
	}
}

func (w *ResponseWriter) Write(d []byte) (int, error) {
	n, err := w.ResponseWriter.Write(d)

	w.Size = n

	return n, err
}

func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}
