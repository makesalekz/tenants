package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/tenants/internal/data"
	"gitlab.calendaria.team/services/tenants/internal/service"
)

func TenantMatchers() (selector.MatchFunc, selector.MatchFunc) {
	return func(ctx context.Context, operation string) bool {
			return operation != "/tenants.v1.Tenants/CreateTenant"
		},
		func(ctx context.Context, operation string) bool {
			return operation == "/tenants.v1.Tenants/CreateTenant"
		}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Bootstrap,
	logger log.Logger,
	jwtp *data.JwtProcessor,
	tenantsService *service.TenantsService,
	membersService *service.MembersService,
) *khttp.Server {
	tenantClaimsMatcher, commonClaimsMatcher := TenantMatchers()

	var opts = []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return jwtp.GetSecret(), nil
				}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256), jwt.WithClaims(func() jwtv4.Claims { return &data.TenantClaims{} })),
			).
				Match(tenantClaimsMatcher).
				Build(),
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return jwtp.GetSecret(), nil
				}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256), jwt.WithClaims(func() jwtv4.Claims { return &jwtv4.RegisteredClaims{} })),
			).
				Match(commonClaimsMatcher).
				Build(),
		),
	}
	if c.Server.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, khttp.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Server.Http.Timeout.AsDuration()))
	}
	srv := khttp.NewServer(opts...)

	v1.RegisterTenantsHTTPServer(srv, tenantsService)
	v1.RegisterMembersHTTPServer(srv, membersService)

	return srv
}
