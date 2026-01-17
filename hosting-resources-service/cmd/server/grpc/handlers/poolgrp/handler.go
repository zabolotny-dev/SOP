package poolgrp

import (
	"context"
	"errors"
	"hosting-resources-service/cmd/server/grpc/handlers/poolgrp/gen"
	"hosting-resources-service/internal/pool"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handlers struct {
	gen.UnimplementedResourcesServer
	poolBus pool.ExtBusiness
}

func New(poolBus pool.ExtBusiness) *Handlers {
	return &Handlers{
		poolBus: poolBus,
	}
}

func (h *Handlers) ConsumeResource(ctx context.Context, req *gen.ConsumeRequest) (*gen.ConsumeReply, error) {
	poolID, err := h.poolBus.ConsumeResource(ctx, pool.Resource{
		CPUCores: int(req.Resource.CpuCores),
		RAMMB:    int(req.Resource.RamMb),
		DiskGB:   int(req.Resource.DiskGb),
		IPCount:  int(req.Resource.IpCount),
	})

	if err != nil {
		if errors.Is(err, pool.ErrValidation) {
			return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
		}
		if errors.Is(err, pool.ErrNotEnoughResources) {
			return nil, status.Errorf(codes.FailedPrecondition, "not enough resources: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "consume resource: %v", err)
	}

	return &gen.ConsumeReply{
		PoolId: poolID.String(),
	}, nil
}

func (h *Handlers) ReturnResource(ctx context.Context, req *gen.ReturnRequest) (*gen.ReturnReply, error) {
	poolID, err := uuid.Parse(req.PoolId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid pool ID: %v", err)
	}

	resource := pool.Resource{
		CPUCores: int(req.Resource.CpuCores),
		RAMMB:    int(req.Resource.RamMb),
		DiskGB:   int(req.Resource.DiskGb),
		IPCount:  int(req.Resource.IpCount),
	}

	err = h.poolBus.ReturnResource(ctx, resource, poolID)

	if err != nil {
		if errors.Is(err, pool.ErrValidation) {
			return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
		}
		if errors.Is(err, pool.ErrPoolNotFound) {
			return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "return resource: %v", err)
	}

	return &gen.ReturnReply{}, nil
}
