package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRole(t *testing.T) {
	claims := &Claims{
		Access: map[string]Access{
			"test-resource": {
				Resource: "test-resource",
				Roles: []string{
					"test-role",
					"other-role",
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
