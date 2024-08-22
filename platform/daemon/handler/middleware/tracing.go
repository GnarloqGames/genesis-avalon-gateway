package middleware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

const (
	tracerName = "daemon"
)

func Tracing(skipPaths ...[]string) func(next http.Handler) http.Handler {
	skip := make(map[string]struct{}, 0)

	if len(skipPaths) > 0 {
		for _, path := range skipPaths[0] {
			skip[path] = struct{}{}
		}
	}

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(skip) > 0 {
				if _, skip := skip[r.URL.Path]; skip {
					next.ServeHTTP(w, r)
					return
				}

				rw := NewResponseWriter(w)
				ctx, span := otel.Tracer(tracerName).Start(r.Context(), "request")
				traceID := span.SpanContext().TraceID().String()
				spanID := span.SpanContext().SpanID().String()

				w.Header().Add("X-Trace-Id", traceID)
				w.Header().Add("X-Span-Id", spanID)

				otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

				next.ServeHTTP(rw, r)

				span.SetAttributes(
					attribute.String("path", r.URL.Path),
					attribute.String("method", r.Method),
					attribute.Int("status", rw.Status),
				)

				code := codes.Ok
				if rw.Status >= 400 {
					code = codes.Error
					span.SetStatus(codes.Error, fmt.Sprintf("%d %s", rw.Status, http.StatusText(rw.Status)))
				}

				span.SetStatus(code, fmt.Sprintf("%d %s", rw.Status, http.StatusText(rw.Status)))

				span.End()
			}
		}

		return http.HandlerFunc(fn)
	}
}
