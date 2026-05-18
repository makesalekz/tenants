//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/makesalekz/tenants/internal/biz"
	"github.com/makesalekz/tenants/internal/conf"
	"github.com/makesalekz/tenants/internal/data"
	"github.com/makesalekz/tenants/internal/server"
	"github.com/makesalekz/tenants/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
