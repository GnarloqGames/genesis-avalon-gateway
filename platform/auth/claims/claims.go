package claims

import (
	"encoding/json"
	"strings"
	"time"
)

type Access struct {
	Resource string
	Roles    []string
}

type Claims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Access    map[string]Access
	ExpiresAt time.Time `json:"exp"`
}

func (c *Claims) MarshalJSON() ([]byte, error) {
	d := make(map[string]any)
	d["sub"] = c.Subject
	d["email"] = c.Email
	d["exp"] = c.ExpiresAt.UnixNano()

	access := make(map[string]map[string]any)

	for _, resource := range c.Access {
		access[resource.Resource] = map[string]any{
			"resource": resource.Resource,
			"roles":    resource.Roles,
		}
	}

	d["resource_access"] = access

	return json.Marshal(d)
}

func (c *Claims) UnmarshalJSON(d []byte) error {
	m := make(map[string]any)
	if err := json.Unmarshal(d, &m); err != nil {
		return err
	}

	c.Subject = m["sub"].(string)
	c.Email = m["email"].(string)
	c.ExpiresAt = time.Unix(0, int64(m["exp"].(float64)))
	c.Access = make(map[string]Access)

	resourceAccess, ok := m["resource_access"].(map[string]any)
	if ok {
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
	}

	return nil
}

func (c *Claims) HasRole(path string) bool {
	segments := strings.SplitN(path, ":", 2)
	if len(segments) < 2 {
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
