package messaging

import (
	"context"
)

type ErrorHandlerFunc func(ctx context.Context, err error, routingKey string)

func LogErrors(handler ErrorHandlerFunc, next MessageHandler) MessageHandler {
	return func(ctx context.Context, body []byte, routingKey string) error {
		err := next(ctx, body, routingKey)
		if err != nil && handler != nil {
			handler(ctx, err, routingKey)
		}
		return err
	}
}
