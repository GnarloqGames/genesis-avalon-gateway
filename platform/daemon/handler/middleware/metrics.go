package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MetricsResponseWriter struct {
	http.ResponseWriter

	Status int
	Size   int
}

func NewMetricsResponseWriter(w http.ResponseWriter) *MetricsResponseWriter {
	return &MetricsResponseWriter{
		ResponseWriter: w,

		Status: 200,
	}
}

func (w *MetricsResponseWriter) Write(d []byte) (int, error) {
	n, err := w.ResponseWriter.Write(d)

	return n, err
}

func (w *MetricsResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *MetricsResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

func Metrics(meter metric.Int64UpDownCounter, skipPaths ...[]string) func(next http.Handler) http.Handler {
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

				mrw := NewMetricsResponseWriter(w)
				next.ServeHTTP(mrw, r)
				meter.Add(r.Context(), 1,
					metric.WithAttributes(attribute.String("path", r.URL.Path)),
					metric.WithAttributes(attribute.String("method", r.Method)),
					metric.WithAttributes(attribute.Int("status", mrw.Status)),
				)
			}
		}

		return http.HandlerFunc(fn)
	}
}
