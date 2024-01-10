package data

import (
	"context"

	rbac_v1 "gitlab.calendaria.team/services/rbac/api/rbac/v1"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/dialer"
)

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

func (r *RbacRemote) GetAssignsClient(ctx context.Context) (rbac_v1.AssignsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, rbac_v1.NewAssignsClient).
		SetEndpoint(r.conf.Discovery.Rbac).
		SetTimeout(r.conf.Discovery.RbacTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *RbacRemote) AssignRole(ctx context.Context, identityId string, tenantId, roleId int64) error {
	client, err := r.GetAssignsClient(ctx)
	if err != nil {
		return tenants_v1.ErrorGrpcConnection("rbac: %s", err.Error())
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
