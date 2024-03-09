package middleware

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Meters struct {
	RequestCounter metric.Int64UpDownCounter
	RequestLength  metric.Float64Histogram
}

func (m Meters) RecordRequest(ctx context.Context, path, method string, status int, measure time.Duration) {
	if m.RequestCounter != nil {
		m.RequestCounter.Add(ctx, 1,
			metric.WithAttributes(attribute.String("path", path)),
			metric.WithAttributes(attribute.String("method", method)),
			metric.WithAttributes(attribute.Int("status", status)),
		)
	}

	if m.RequestLength != nil {
		m.RequestLength.Record(ctx, float64(measure.Milliseconds()),
			metric.WithAttributes(attribute.String("path", path)),
			metric.WithAttributes(attribute.String("method", method)),
			metric.WithAttributes(attribute.Int("status", status)),
		)
	}
}

func Metrics(meter Meters, skipPaths ...[]string) func(next http.Handler) http.Handler {
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

				mrw := NewResponseWriter(w)
				start := time.Now()

				next.ServeHTTP(mrw, r)

				measure := time.Since(start)
				meter.RecordRequest(r.Context(), r.URL.Path, r.Method, mrw.Status, measure)
			}
		}

		return http.HandlerFunc(fn)
	}
}
