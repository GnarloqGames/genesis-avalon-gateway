package handler

import (
	"log/slog"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	http.ResponseWriter

	Status int
	Size   int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,

		Status: 200,
		Size:   0,
	}
}

func (w *LoggingResponseWriter) Write(d []byte) (int, error) {
	n, err := w.ResponseWriter.Write(d)

	w.Size = n

	return n, err
}

func (w *LoggingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *LoggingResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(skipPaths ...[]string) func(next http.Handler) http.Handler {
	skip := make(map[string]struct{}, 0)

	if len(skipPaths) > 0 {
		for _, path := range skipPaths[0] {
			skip[path] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(skip) > 0 {
				if _, skip := skip[r.URL.Path]; skip {
					next.ServeHTTP(w, r)
					return
				}
			}

			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			slog.Info("http request",
				"remote_addr", r.RemoteAddr,
				"remote_user", r.URL.User.Username(),
				"time_local", time.Now().Format(time.RFC3339),
				"request_method", r.Method,
				"request_path", r.URL.Path,
				"request_protocol", r.Proto,
				"status", lrw.Status,
				"body_bytes_sent", lrw.Size,
				"http_referer", r.Referer(),
				"http_user_agent", r.UserAgent(),
			)
		}

		return http.HandlerFunc(fn)
	}
}
