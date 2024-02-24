package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setIDTokenClaims(idToken *oidc.IDToken, claims []byte) {
	pointerVal := reflect.ValueOf(idToken)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("claims")
	ptr := unsafe.Pointer(member.UnsafeAddr())
	realPtr := (*[]byte)(ptr)
	*realPtr = claims
}

func TestInjectClaims(t *testing.T) {
	subject := "196176fd-6e54-49c2-9e49-eb81406c68d5"
	now := time.Now().Add(5 * time.Minute)

	patches := gomonkey.ApplyFunc(Verify, func(ctx context.Context, accessToken string) (*oidc.IDToken, error) {
		claims := fmt.Sprintf(`{
"email": "test@test.com",
"email_verified": true,
"exp": %d,
"preferred_username": "testing",
"resource_access": {
    "test-resource": {
        "roles": [
            "test-role"
        ]
    }
},
"sub": "%s"
}`,
			now.Unix(), subject)

		token := &oidc.IDToken{
			Subject: subject,
			Expiry:  now,
		}
		setIDTokenClaims(token, []byte(claims))

		return token, nil
	})
	defer patches.Reset()

	ctx, err := injectClaims(context.Background(), "foo")
	assert.NoError(t, err)

	claims, ok := ctx.Value(ClaimsContext).(*Claims)
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
	patches := gomonkey.ApplyFunc(Verify, func(ctx context.Context, accessToken string) (*oidc.IDToken, error) {
		claims := fmt.Sprintf(`{
"email": "test@test.com",
"email_verified": true,
"exp": %d,
"preferred_username": "testing",
"resource_access": {
    "test-resource": {
        "roles": [
            "test-role"
        ]
    }
},
"sub": "%s"
}`,
			now.Unix(), subject)

		token := &oidc.IDToken{
			Subject: subject,
			Expiry:  now,
		}
		setIDTokenClaims(token, []byte(claims))

		return token, nil
	})
	defer patches.Reset()

	fn := Middleware()

	var ctx context.Context

	inner := func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}

	outer := fn(http.HandlerFunc(inner))

	success := func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "", nil)
		req.Header.Add("Authorization", "Bearer asd")
		assert.NoError(t, err)

		mrw := &MockResponseWriter{}
		outer.ServeHTTP(mrw, req)

		claims, ok := ctx.Value(ClaimsContext).(*Claims)
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
