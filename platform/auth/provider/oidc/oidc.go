package oidc

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
)

type OidcProvider struct {
	*oidc.Provider

	clientID string
}

func New(provider, clientID string) (*OidcProvider, error) {
	p, err := oidc.NewProvider(context.Background(), provider)
	if err != nil {
		return nil, err
	}

	return &OidcProvider{
		Provider: p,
		clientID: clientID,
	}, nil
}

func (m *OidcProvider) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	verifier := m.Provider.Verifier(&oidc.Config{ClientID: m.clientID})

	return verifier.Verify(ctx, rawIDToken)
}
