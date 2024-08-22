package mockverifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/claims"
	"github.com/coreos/go-oidc/v3/oidc"
)

type MockProvider struct {
	expected map[string]*claims.Claims
}

func (m *MockProvider) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	exp, ok := m.expected[rawIDToken]
	if !ok {
		return nil, fmt.Errorf("invalid id token")
	}

	body := bytes.NewBufferString("")

	encoder := json.NewEncoder(body)
	if err := encoder.Encode(exp); err != nil {
		return nil, err
	}

	token := &oidc.IDToken{
		Subject: exp.Subject,
		Expiry:  exp.ExpiresAt,
	}
	setIDTokenClaims(token, body.Bytes())

	return token, nil
}

type Expectation struct {
	Token  string
	Claims *claims.Claims
}

func New(expectations ...Expectation) *MockProvider {
	p := &MockProvider{
		expected: make(map[string]*claims.Claims),
	}

	for _, exp := range expectations {
		p.expected[exp.Token] = exp.Claims
	}

	return p
}

func setIDTokenClaims(idToken *oidc.IDToken, claims []byte) {
	pointerVal := reflect.ValueOf(idToken)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("claims")
	ptr := unsafe.Pointer(member.UnsafeAddr())
	realPtr := (*[]byte)(ptr)
	*realPtr = claims
}
