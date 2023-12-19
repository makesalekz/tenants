package data

import (
	"context"
	"strconv"
	"time"

	rbac_v1 "gitlab.calendaria.team/services/rbac/api/rbac/v1"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/dialer"
	"gitlab.calendaria.team/services/utils/v1/jwt"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

const TOKEN_DURATION = 1 * time.Minute

type RbacRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewRbacRemote(d *dialer.Dialer, conf *conf.Bootstrap) (*RbacRemote, error) {
	return &RbacRemote{
		dialer: d,
		conf:   conf,
	}, nil
}

func (r *RbacRemote) GetAssignsClient(ctx context.Context, claims *jwt.TenantClaims) (rbac_v1.AssignsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, rbac_v1.NewAssignsClient).
		SetEndpoint(r.conf.Discovery.Rbac).
		SetTimeout(r.conf.Discovery.RbacTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *RbacRemote) AssignRole(ctx context.Context, identityId string, tenantId, ownerId, roleId int64) error {
	claims := &jwt.TenantClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Issuer:    "iam",
			Audience:  jwtv4.ClaimStrings{"tenant"},
			Subject:   strconv.FormatInt(ownerId, 10),
			IssuedAt:  jwtv4.NewNumericDate(time.Now()),
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(TOKEN_DURATION)),
		},
		TenantId: tenantId,
	}

	client, err := r.GetAssignsClient(ctx, claims)
	if err != nil {
		return tenants_v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	_, err = client.AssignRole(ctx, &rbac_v1.AssignRoleRequest{
		IdentityId: identityId,
		RoleId:     roleId,
	})
	if err != nil {
		return err
	}

	return nil
}
