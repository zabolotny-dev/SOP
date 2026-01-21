package auth

import (
	"context"
)

type ctxKey int

const claimsKey ctxKey = 0

func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}, ErrUnauthorized
	}
	return claims, nil
}
