package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Sender interface {
	Send(ctx context.Context, userID uuid.UUID, event Event) error
}

type Notifier struct {
	senders []Sender
}

func New(senders ...Sender) *Notifier {
	return &Notifier{senders: senders}
}

func (n *Notifier) Notify(ctx context.Context, userID uuid.UUID, event Event) error {
	var errs []error

	for _, sender := range n.senders {
		if err := sender.Send(ctx, userID, event); err != nil {
			errs = append(errs, fmt.Errorf("sender %T failed: %w", sender, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification completed with errors: %v", errs)
	}
	return nil
}
