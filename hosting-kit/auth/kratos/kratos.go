package kratos

import (
	"context"
	"fmt"
	"hosting-kit/auth"
	"hosting-kit/otel"

	"github.com/google/uuid"
	ory "github.com/ory/kratos-client-go"
)

type client struct {
	kratos *ory.APIClient
}

func New(kratosURL string) auth.Client {
	config := ory.NewConfiguration()
	config.Servers = []ory.ServerConfiguration{
		{URL: kratosURL},
	}

	return &client{
		kratos: ory.NewAPIClient(config),
	}
}

func (c *client) Authenticate(ctx context.Context, token string) (auth.Claims, error) {
	ctx, span := otel.AddSpan(ctx, "auth.kratos.authenticate")
	defer span.End()

	if token == "" {
		return auth.Claims{}, auth.ErrUnauthorized
	}

	session, resp, err := c.kratos.FrontendAPI.
		ToSession(ctx).
		Cookie(token).
		Execute()

	if err != nil || resp.StatusCode != 200 {
		return auth.Claims{}, fmt.Errorf("%w: invalid session", auth.ErrUnauthorized)
	}

	traits, ok := session.Identity.Traits.(map[string]interface{})
	if !ok {
		return auth.Claims{}, fmt.Errorf("%w: invalid traits", auth.ErrUnauthorized)
	}

	email, _ := traits["email"].(string)
	name, _ := traits["name"].(string)

	isAdmin := false
	if session.Identity.MetadataPublic != nil {
		if meta, ok := session.Identity.MetadataPublic.(map[string]interface{}); ok {
			if val, ok := meta["is_admin"].(bool); ok {
				isAdmin = val
			}
		}
	}

	claims := auth.Claims{
		UserID:  uuid.MustParse(session.Identity.Id),
		Email:   email,
		Name:    name,
		IsAdmin: isAdmin,
	}

	return claims, nil
}
