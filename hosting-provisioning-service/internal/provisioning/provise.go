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

type Extension func(ExtBusiness) ExtBusiness

type Notifier interface {
	NotifySuccess(ctx context.Context, serverID uuid.UUID, res Result) error
	NotifyFailure(ctx context.Context, serverID uuid.UUID, reason string, failedAt time.Time) error
}

type Business struct {
	provisioningTime time.Duration
	notifier         Notifier
	extensions       []Extension
}

type ExtBusiness interface {
	GenerateIP(ctx context.Context, serverID uuid.UUID) error
}

func NewBusiness(provisioningTime time.Duration, notifier Notifier, extensions ...Extension) ExtBusiness {
	b := &Business{
		provisioningTime: provisioningTime,
		notifier:         notifier,
		extensions:       extensions,
	}

	extBus := ExtBusiness(b)

	for i := len(extensions) - 1; i >= 0; i-- {
		ext := extensions[i]
		if ext != nil {
			extBus = ext(extBus)
		}
	}

	return extBus
}

type Result struct {
	IP            string
	ProvisionedAt time.Time
}

func (ps *Business) GenerateIP(ctx context.Context, serverID uuid.UUID) error {

	select {
	case <-time.After(ps.provisioningTime):
	case <-ctx.Done():
		return fmt.Errorf("generateip: provisioning cancelled for server %s", serverID)
	}

	if rand.Intn(10) < 2 {
		if err := ps.notifier.NotifyFailure(ctx, serverID, "IP generation failed", time.Now().UTC()); err != nil {
			return fmt.Errorf("generateip: %w", err)
		}
		return nil
	}

	ip := fmt.Sprintf("10.0.0.%d", rand.Intn(255))
	if err := ps.notifier.NotifySuccess(ctx, serverID, Result{
		IP:            ip,
		ProvisionedAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("generateip: %w", err)
	}
	return nil
}
