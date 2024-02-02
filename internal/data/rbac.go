package data

import (
	"context"

	rbac_v1 "gitlab.calendaria.team/services/rbac/api/rbac/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/config"
	jwtp "gitlab.calendaria.team/services/utils/v1/jwt"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type RbacRemote struct {
	dialer *dialer.Dialer
}

func NewRbacRemote(
	conf *conf.Bootstrap,
	c *config.Config,
	jwt *jwtp.JwtProcessor,
) (*RbacRemote, error) {
	dialer, err := dialer.NewServiceDialer(c, jwt, "rbac", conf.Discovery.Rbac)
	if err != nil {
		return nil, err
	}

	return &RbacRemote{
		dialer: dialer,
	}, nil
}

func (r *RbacRemote) getAssignsClient(ctx context.Context) (rbac_v1.AssignsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to rbac: %s", err.Error())
	}

	return rbac_v1.NewAssignsClient(conn), nil
}

func (r *RbacRemote) AssignRole(ctx context.Context, identityId string, tenantId, roleId int64) error {
	client, err := r.getAssignsClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.AssignRole(
		ctx,
		&rbac_v1.AssignRoleRequest{
			TenantId:   tenantId,
			IdentityId: identityId,
			RoleId:     roleId,
		})
	if err != nil {
		return err
	}

	return nil
}
