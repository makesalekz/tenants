package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/tenants/internal/service"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/tenants.v1.Invites/ShownInvite"] = struct{}{}
	whiteList["/tenants.v1.Invites/DeclineInvite"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Bootstrap,
	logger log.Logger,
	jwtp *jwt.JwtProcessor,
	tenantsService *service.TenantsService,
	membersService *service.MembersService,
	invitesService *service.InvitesService,
) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			selector.Server(
				kjwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return jwtp.GetSecret(), nil
				}, kjwt.WithSigningMethod(jwtv4.SigningMethodHS256), kjwt.WithClaims(func() jwtv4.Claims { return &jwt.TenantClaims{} })),
			).
				Match(NewWhiteListMatcher()).
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
	v1.RegisterInvitesHTTPServer(srv, invitesService)

	return srv
}
