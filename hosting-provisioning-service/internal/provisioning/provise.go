package provisioning

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var ErrIpGenerationFailed = errors.New("generation failed")

type Notifier interface {
	NotifySuccess(ctx context.Context, serverID uuid.UUID, res Result) error
	NotifyFailure(ctx context.Context, serverID uuid.UUID, reason string, failedAt time.Time) error
}

type Business struct {
	provisioningTime time.Duration
	notifier         Notifier
}

func NewBusiness(provisioningTime time.Duration, notifier Notifier) *Business {
	return &Business{
		provisioningTime: provisioningTime,
		notifier:         notifier,
	}
}

type Result struct {
	IP            string
	ProvisionedAt time.Time
}

func (ps *Business) GenerateIP(ctx context.Context, serverID uuid.UUID) error {

	select {
	case <-time.After(ps.provisioningTime):
	case <-ctx.Done():
		return fmt.Errorf("provisioning cancelled for server %s", serverID)
	}

	if rand.Intn(10) < 2 {
		if err := ps.notifier.NotifyFailure(ctx, serverID, "IP generation failed", time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to notify failure: %w", err)
		}
		return nil
	}

	ip := fmt.Sprintf("10.0.0.%d", rand.Intn(255))
	if err := ps.notifier.NotifySuccess(ctx, serverID, Result{
		IP:            ip,
		ProvisionedAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("failed to notify success: %w", err)
	}
	return nil
}
