package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(skipPaths ...[]string) func(next http.Handler) http.Handler {
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

			lrw := NewResponseWriter(w)
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
