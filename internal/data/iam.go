package data

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
)

type IamRemote struct {
	dialer *Dialer
}

func NewIamRemote(dialer *Dialer) (*IamRemote, error) {
	return &IamRemote{
		dialer: dialer,
	}, nil
}

// GetUser returns userShort from iam service by userId.
func (r *IamRemote) GetUser(ctx context.Context, userId int64) (*iam_v1.UserShort, error) {
	usersClient, err := r.dialer.Users(ctx)
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
func (r *IamRemote) GetUsers(ctx context.Context, usersIds []int64, emails []string) ([]*iam_v1.UserShort, error) {
	if len(usersIds) == 0 && len(emails) == 0 {
		return nil, nil
	}

	usersClient, err := r.dialer.Users(ctx)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	reply, err := usersClient.GetUsers(ctx, &iam_v1.GetUsersRequest{Ids: usersIds, Emails: emails})
	if err != nil {
		return nil, tenants_v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	return reply.Users, nil
}
