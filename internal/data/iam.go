package data

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/utils/v4/dialer"

	"github.com/go-kratos/kratos/v2/log"
)

type IamRemote struct {
	log    *log.Helper
	dialer dialer.IDialer
}

func NewIamRemote(
	logger log.Logger,
	conf *conf.Bootstrap,
	dm dialer.IDialerManager,
) (IIamRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("iam", conf.GetDiscovery().GetIam())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &IamRemote{
		log:    log.NewHelper(log.With(logger, "module", "data/iam")),
		dialer: dialer,
	}, cleanup, nil
}

func (r *IamRemote) getUsersClient(ctx context.Context) (iam_v1.UsersClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to iam: %s", err.Error())
	}

	return iam_v1.NewUsersClient(conn), nil
}

// GetUser returns userShort from iam service by userId.
func (r *IamRemote) GetUser(ctx context.Context, userID int64) (*iam_v1.UserShort, error) {
	usersClient, err := r.getUsersClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := usersClient.GetUser(ctx, &iam_v1.GetUserRequest{UserId: userID})
	if err != nil {
		if iam_v1.IsUserNotFound(err) {
			return nil, v1.ErrorNotFound("user not found")
		}
		return nil, v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply.GetUser(), nil
}

// GetUsers returns userShorts map from iam service by mapUsersIds.
func (r *IamRemote) GetUsers(ctx context.Context, req *iam_v1.GetUsersRequest) (*iam_v1.UsersReply, error) {
	usersClient, err := r.getUsersClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := usersClient.GetUsers(ctx, req)
	if err != nil {
		return nil, v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply, nil
}

// GetUsers returns userShorts map from iam service by mapUsersIds.
func (r *IamRemote) ListUsers(ctx context.Context, req *iam_v1.ListUsersRequest) (*iam_v1.UsersReply, error) {
	usersClient, err := r.getUsersClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := usersClient.ListUsers(ctx, req)
	if err != nil {
		return nil, v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply, nil
}
