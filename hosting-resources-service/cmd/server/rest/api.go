package rest

import (
	"hosting-resources-service/cmd/server/rest/handlers/poolgrp"
	"hosting-resources-service/cmd/server/rest/handlers/rootgrp"
	"hosting-resources-service/internal/pool"
)

type API struct {
	*poolgrp.PoolHandlers
	*rootgrp.RootHandlers
}

func New(poolBus pool.ExtBusiness, prefix string) *API {
	return &API{
		PoolHandlers: poolgrp.New(poolBus, prefix),
		RootHandlers: rootgrp.New(prefix),
	}
}
