package auth

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/viper"
)

var provider *oidc.Provider

type Access struct {
	Resource string
	Roles    []string
}

type Claims struct {
	Subject   string
	Email     string
	Access    map[string]Access
	ExpiresAt time.Time
}

type authContextKey string

const ClaimsContext authContextKey = "claims"

func (c *Claims) UnmarshalJSON(d []byte) error {
	m := make(map[string]any)
	if err := json.Unmarshal(d, &m); err != nil {
		return err
	}

	c.Subject = m["sub"].(string)
	c.Email = m["email"].(string)
	c.ExpiresAt = time.Unix(int64(m["exp"].(float64)), 0)
	c.Access = make(map[string]Access)

	resourceAccess := m["resource_access"].(map[string]any)
	for resource, access := range resourceAccess {
		roles := make([]string, 0)
		resourceRoles := access.(map[string]any)["roles"]

		for _, role := range resourceRoles.([]any) {
			roles = append(roles, role.(string))
		}

		c.Access[resource] = Access{
			Resource: resource,
			Roles:    roles,
		}
	}

	return nil
}

func InitProvider() error {
	if provider == nil {
		p, err := oidc.NewProvider(context.Background(), viper.GetString("oidc-provider"))
		if err != nil {
			return err
		}

		provider = p
	}

	return nil
}

func Verify(ctx context.Context, accessToken string) (*oidc.IDToken, error) {
	verifier := provider.Verifier(&oidc.Config{ClientID: viper.GetString("oidc-client-id")})

	return verifier.Verify(ctx, accessToken)
}

func (c *Claims) HasRole(path string) bool {
	segments := strings.Split(path, ":")
	if len(segments) != 2 {
		return false
	}

	access, resourceExists := c.Access[segments[0]]
	if !resourceExists {
		return false
	}

	for _, role := range access.Roles {
		if role == segments[1] {
			return true
		}
	}

	return false
}
