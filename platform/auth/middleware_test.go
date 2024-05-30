package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/claims"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/provider/mockverifier"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInjectClaims(t *testing.T) {
	subject := "196176fd-6e54-49c2-9e49-eb81406c68d5"
	now := time.Now().Add(5 * time.Minute)

	verifier := mockverifier.New(mockverifier.Expectation{
		Token: "foo",
		Claims: &claims.Claims{
			Subject:   subject,
			ExpiresAt: now,
			Access: map[string]claims.Access{
				"test-resource": {
					Resource: "test-resource",
					Roles: []string{
						"test-role",
					},
				},
			},
		},
	})

	ctx, err := injectClaims(context.Background(), verifier, "foo")
	assert.NoError(t, err)

	claims, ok := ctx.Value(ClaimsContext).(*claims.Claims)
	assert.True(t, ok)
	assert.NotNil(t, claims)
	assert.Equal(t, now.Unix(), claims.ExpiresAt.Unix())
	assert.Equal(t, subject, claims.Subject)
}

type MockResponseWriter struct {
	Status int
}

func (m *MockResponseWriter) Header() http.Header         { return http.Header{} }
func (m *MockResponseWriter) Write(d []byte) (int, error) { slog.Info(string(d)); return len(d), nil }
func (m *MockResponseWriter) WriteHeader(status int)      { m.Status = status }

func TestMiddleware(t *testing.T) {
	subject := uuid.New().String()
	now := time.Now().Add(5 * time.Minute)

	token := uuid.New()

	verifier := mockverifier.New(mockverifier.Expectation{
		Token: token.String(),
		Claims: &claims.Claims{
			Subject:   subject,
			ExpiresAt: now,
			Access: map[string]claims.Access{
				"test-resource": {
					Resource: "test-resource",
					Roles: []string{
						"test-role",
					},
				},
			},
		},
	})

	fn := Middleware(verifier)

	var ctx context.Context

	inner := func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}

	outer := fn(http.HandlerFunc(inner))

	success := func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.String()))
		assert.NoError(t, err)

		mrw := &MockResponseWriter{}
		outer.ServeHTTP(mrw, req)

		claims, ok := ctx.Value(ClaimsContext).(*claims.Claims)
		assert.True(t, ok)
		assert.NotNil(t, claims)
		assert.Equal(t, now.Unix(), claims.ExpiresAt.Unix())
		assert.Equal(t, subject, claims.Subject)
	}

	emptyAuth := func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "", nil)
		assert.NoError(t, err)

		mrw := &MockResponseWriter{}
		outer.ServeHTTP(mrw, req)

		assert.Equal(t, http.StatusUnauthorized, mrw.Status)
	}

	t.Run("happy-path", success)
	t.Run("empty-auth", emptyAuth)
}
