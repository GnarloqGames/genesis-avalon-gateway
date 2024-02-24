package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

func Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if accessToken == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx, err := injectClaims(r.Context(), accessToken)
			if err != nil {
				slog.Error("failed to inject claims into context", "error", err, "access_token", accessToken)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

				return
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func injectClaims(ctx context.Context, accessToken string) (context.Context, error) {
	idToken, err := Verify(ctx, accessToken)

	if err != nil {
		return ctx, fmt.Errorf("failed to verify access token: %w", err)
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return ctx, fmt.Errorf("failed to extract claims from token: %w", err)
	}

	newCtx := context.WithValue(ctx, ClaimsContext, &claims)

	return newCtx, nil
}
