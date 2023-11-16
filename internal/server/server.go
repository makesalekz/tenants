package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)

func TenantMatchers() (selector.MatchFunc, selector.MatchFunc) {
	return func(ctx context.Context, operation string) bool {
			return operation != "/tenants.v1.Tenants/CreateTenant"
		},
		func(ctx context.Context, operation string) bool {
			return operation == "/tenants.v1.Tenants/CreateTenant"
		}
}
