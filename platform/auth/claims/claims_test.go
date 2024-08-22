package claims

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
)

func TestHasRole(t *testing.T) {
	claims := &Claims{
		Access: map[string]Access{
			"test-resource": {
				Resource: "test-resource",
				Roles: []string{
					"test-role",
					"other-role",
					"nested:role",
				},
			},
		},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{
			path:     "test-resource:test-role",
			expected: true,
		},
		{
			path:     "test-resource:bogus",
			expected: false,
		},
		{
			path:     "test-resource:nested:role",
			expected: true,
		},
		{
			path:     "non-existent:resource",
			expected: false,
		},
		{
			path:     "wrong-number",
			expected: false,
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			actual := claims.HasRole(tt.path)
			assert.Equal(t, tt.expected, actual)
		}

		t.Run(tt.path, tf)
	}
}

func TestMarshalJSON(t *testing.T) {
	subject := "test"
	expires := time.Now()
	email := "bogus@test.com"
	access := Access{
		Resource: "test-resource",
		Roles: []string{
			"test-role",
			"other-role",
		},
	}
	claims := &Claims{
		Subject:   subject,
		ExpiresAt: expires,
		Email:     email,
		Access: map[string]Access{
			"test-resource": access,
		},
	}

	raw, err := json.Marshal(claims)
	require.NoError(t, err)

	var newClaims *Claims
	err = json.Unmarshal(raw, &newClaims)
	require.NoError(t, err)

	require.Equal(t, claims.Access, newClaims.Access)
	require.Equal(t, claims.Subject, newClaims.Subject)
	require.Equal(t, claims.Email, newClaims.Email)
	require.WithinRange(t, newClaims.ExpiresAt, claims.ExpiresAt.Add(-500*time.Millisecond), claims.ExpiresAt.Add(500*time.Millisecond))
}

func TestUnmarshalInvalid(t *testing.T) {
	raw := []byte("not json")
	var claims *Claims

	err := claims.UnmarshalJSON(raw)
	require.Error(t, err)
}
