package data

import (
	"context"

	rbac_v1 "gitlab.calendaria.team/services/rbac/api/rbac/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v2/dialer"

	"github.com/go-kratos/kratos/v2/log"
)

type RbacRemote struct {
	log    *log.Helper
	dialer dialer.IDialer
}

func NewRbacRemote(
	logger log.Logger,
	conf *conf.Bootstrap,
	dm dialer.IDialerManager,
) (*RbacRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("rbac", conf.Discovery.Rbac)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &RbacRemote{
		log:    log.NewHelper(log.With(logger, "module", "data/rbac")),
		dialer: dialer,
	}, cleanup, nil
}

func (r *RbacRemote) getAssignsClient(ctx context.Context) (rbac_v1.AssignsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to rbac: %s", err.Error())
	}

	return rbac_v1.NewAssignsClient(conn), nil
}

func (r *RbacRemote) AssignRole(ctx context.Context, identityId string, roleId int64) error {
	client, err := r.getAssignsClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.AssignRole(
		ctx,
		&rbac_v1.AssignRoleRequest{
			IdentityId: identityId,
			RoleId:     roleId,
		})
	if err != nil {
		return err
	}

	return nil
}
