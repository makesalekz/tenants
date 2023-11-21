package server

import (
	"github.com/go-kratos/kratos/v2/log"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/tenants/internal/service"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Bootstrap,
	logger log.Logger,
	jwtp *jwt.JwtProcessor,
	tenantsService *service.TenantsService,
	membersService *service.MembersService,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			kjwt.Server(func(token *jwtv4.Token) (interface{}, error) {
				return jwtp.GetSecret(), nil
			}, kjwt.WithSigningMethod(jwtv4.SigningMethodHS256), kjwt.WithClaims(func() jwtv4.Claims { return &jwt.TenantClaims{} })),
		),
	}
	if c.Server.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Server.Grpc.Network))
	}
	if c.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Server.Grpc.Addr))
	}
	if c.Server.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Server.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	tenants_v1.RegisterTenantsServer(srv, tenantsService)
	tenants_v1.RegisterMembersServer(srv, membersService)

	return srv
}
