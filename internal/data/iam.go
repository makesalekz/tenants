package data

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/dialer"
)

type IamRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewIamRemote(d *dialer.Dialer, conf *conf.Bootstrap) (*IamRemote, error) {
	return &IamRemote{
		dialer: d,
		conf:   conf,
	}, nil
}

func (r *IamRemote) GetUsersClient(ctx context.Context) (iam_v1.UsersClient, error) {
	return dialer.NewDialerBuilder(r.dialer, iam_v1.NewUsersClient).
		SetEndpoint(r.conf.Discovery.Iam).
		SetTimeout(r.conf.Discovery.IamTimeout.AsDuration()).
		Conn(ctx, nil)
}

// GetUser returns userShort from iam service by userId.
func (r *IamRemote) GetUser(ctx context.Context, userId int64) (*iam_v1.UserShort, error) {
	usersClient, err := r.GetUsersClient(ctx)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	reply, err := usersClient.GetUser(ctx, &iam_v1.GetUserRequest{UserId: userId})
	if err != nil {
		if iam_v1.IsUserNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("user not found")
		}
		return nil, tenants_v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply.User, nil
}

// GetUsers returns userShorts map from iam service by mapUsersIds.
func (r *IamRemote) GetUsers(ctx context.Context, req *iam_v1.GetUsersRequest) (*iam_v1.UsersReply, error) {
	usersClient, err := r.GetUsersClient(ctx)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	reply, err := usersClient.GetUsers(ctx, req)
	if err != nil {
		return nil, tenants_v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply, nil
}

// GetUsers returns userShorts map from iam service by mapUsersIds.
func (r *IamRemote) ListUsers(ctx context.Context, req *iam_v1.ListUsersRequest) (*iam_v1.UsersReply, error) {
	usersClient, err := r.GetUsersClient(ctx)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	reply, err := usersClient.ListUsers(ctx, req)
	if err != nil {
		return nil, tenants_v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply, nil
}
