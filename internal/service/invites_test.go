package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	"gitlab.calendaria.team/services/tenants/internal/data/mock"
	"gitlab.calendaria.team/services/tenants/internal/service"
	nats_mock "gitlab.calendaria.team/services/utils/v1/nats/mock"
	"gitlab.calendaria.team/services/utils/v2/auth"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"
	"gitlab.calendaria.team/services/utils/v2/zap"
)

func TestInvitesCreate(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	queue := nats_mock.NewMockIQueueManager(ctrl)
	invitesRepo := mock.NewMockInvitesRepo(ctrl)
	tenantsRepo := mock.NewMockTenantsRepo(ctrl)
	iamRemote := mock.NewMockIIamRemote(ctrl)
	rbacRemote := mock.NewMockIRbacRemote(ctrl)
	config := mock.NewMockConfig(ctrl)

	invitesUsecase, err := biz.NewInvitesUsecase(
		logger, tenantsRepo, invitesRepo, iamRemote, rbacRemote, queue, config,
	)
	require.NoError(t, err)

	invitesService := service.NewInvitesService(nil, invitesUsecase)

	ctx := context.Background()

	var tenantID int64 = 12
	var actorID int64 = 332
	appID := "pms"
	emails := []string{"email1", "email2"}
	inviteCode := u_uuid.NewFromActorID(actorID)
	invitesDto := &data.InvitesDTO{
		Emails:     emails,
		Lang:       "en",
		RoleID:     5, // project viewer
		Resource:   "project",
		ResourceID: 223,
	}
	users := []*iam_v1.UserShort{
		{
			Id:    1,
			Name:  "name1",
			Email: "email1",
		},
		{
			Id:    2,
			Name:  "name2",
			Email: "email2",
		},
	}

	inviteDtos := []data.InviteDto{
		data.InviteDto{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     5,
			Resource:   "project",
			ResourceID: 223,
		},
		data.InviteDto{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     5,
			Resource:   "project",
			ResourceID: 223,
		},
	}

	entInvites := []*ent.Invite{
		&ent.Invite{
			ID:         123,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[0],
			UserID:     &users[0].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     5,
			Resource:   "project",
			ResourceID: 123,
			Edges:      ent.InviteEdges{},
		},
		&ent.Invite{
			ID:         223,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[1],
			UserID:     &users[1].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     5,
			Resource:   "project",
			ResourceID: 123,
			Edges:      ent.InviteEdges{},
		},
	}

	expectedInvites := service.ReplyInvites(
		[]biz.InviteItem{
			{
				Invite: entInvites[0],
				User:   users[0],
			},
			{
				Invite: entInvites[1],
				User:   users[1],
			},
		},
	)

	iamRemote.EXPECT().GetUsers(
		gomock.Any(), &iam_v1.GetUsersRequest{Emails: emails},
	).Return(&iam_v1.UsersReply{Users: users}, nil)
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, inviteDtos).Return(entInvites, nil)
	tenantsRepo.EXPECT().GetTenant(gomock.Any(), tenantID).Return(
		&ent.Tenant{}, nil,
	).Do(func() { time.Sleep(time.Second) }).AnyTimes() // this method is used in goroutine, so we need to wait

	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)

	invites, err := invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     invitesDto.RoleID,
			Resource:   invitesDto.Resource,
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.NoError(t, err)

	require.Equal(t, expectedInvites, invites.GetInvites())
}

func TestInvitesCreateWithoutResource(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	queue := nats_mock.NewMockIQueueManager(ctrl)
	invitesRepo := mock.NewMockInvitesRepo(ctrl)
	tenantsRepo := mock.NewMockTenantsRepo(ctrl)
	iamRemote := mock.NewMockIIamRemote(ctrl)
	rbacRemote := mock.NewMockIRbacRemote(ctrl)
	config := mock.NewMockConfig(ctrl)

	invitesUsecase, err := biz.NewInvitesUsecase(
		logger, tenantsRepo, invitesRepo, iamRemote, rbacRemote, queue, config,
	)
	require.NoError(t, err)

	invitesService := service.NewInvitesService(nil, invitesUsecase)

	ctx := context.Background()

	var tenantID int64 = 12
	var actorID int64 = 332
	appID := "pms"
	emails := []string{"email1", "email2"}
	inviteCode := u_uuid.NewFromActorID(actorID)
	invitesDto := &data.InvitesDTO{
		Emails:     emails,
		Lang:       "en",
		RoleID:     5, // project viewer
		Resource:   "",
		ResourceID: 0,
	}
	users := []*iam_v1.UserShort{
		{
			Id:    1,
			Name:  "name1",
			Email: "email1",
		},
		{
			Id:    2,
			Name:  "name2",
			Email: "email2",
		},
	}

	inviteDtos := []data.InviteDto{
		data.InviteDto{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
		},
		data.InviteDto{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
		},
	}

	entInvites := []*ent.Invite{
		&ent.Invite{
			ID:         123,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[0],
			UserID:     &users[0].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
			Edges:      ent.InviteEdges{},
		},
		&ent.Invite{
			ID:         223,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[1],
			UserID:     &users[1].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
			Edges:      ent.InviteEdges{},
		},
	}

	expectedInvites := service.ReplyInvites(
		[]biz.InviteItem{
			{
				Invite: entInvites[0],
				User:   users[0],
			},
			{
				Invite: entInvites[1],
				User:   users[1],
			},
		},
	)

	iamRemote.EXPECT().GetUsers(
		gomock.Any(), &iam_v1.GetUsersRequest{Emails: emails},
	).Return(&iam_v1.UsersReply{Users: users}, nil)
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, inviteDtos).Return(entInvites, nil)
	tenantsRepo.EXPECT().GetTenant(gomock.Any(), tenantID).Return(
		&ent.Tenant{}, nil,
	).Do(func() { time.Sleep(time.Second) }).AnyTimes() // this method is used in goroutine, so we need to wait

	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)

	invites, err := invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     invitesDto.RoleID,
			Resource:   invitesDto.Resource,
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.NoError(t, err)

	require.Equal(t, expectedInvites, invites.GetInvites())
}

func TestInvitesCreateWithoutRole(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	queue := nats_mock.NewMockIQueueManager(ctrl)
	invitesRepo := mock.NewMockInvitesRepo(ctrl)
	tenantsRepo := mock.NewMockTenantsRepo(ctrl)
	iamRemote := mock.NewMockIIamRemote(ctrl)
	rbacRemote := mock.NewMockIRbacRemote(ctrl)
	config := mock.NewMockConfig(ctrl)

	invitesUsecase, err := biz.NewInvitesUsecase(
		logger, tenantsRepo, invitesRepo, iamRemote, rbacRemote, queue, config,
	)
	require.NoError(t, err)

	invitesService := service.NewInvitesService(nil, invitesUsecase)

	ctx := context.Background()

	var tenantID int64 = 12
	var actorID int64 = 332
	appID := "pms"
	emails := []string{"email1", "email2"}
	inviteCode := u_uuid.NewFromActorID(actorID)
	invitesDto := &data.InvitesDTO{
		Emails:     emails,
		Lang:       "en",
		RoleID:     0,
		Resource:   "",
		ResourceID: 0,
	}
	users := []*iam_v1.UserShort{
		{
			Id:    1,
			Name:  "name1",
			Email: "email1",
		},
		{
			Id:    2,
			Name:  "name2",
			Email: "email2",
		},
	}

	inviteDtos := []data.InviteDto{
		data.InviteDto{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
		},
		data.InviteDto{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
		},
	}

	entInvites := []*ent.Invite{
		&ent.Invite{
			ID:         123,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[0],
			UserID:     &users[0].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
			Edges:      ent.InviteEdges{},
		},
		&ent.Invite{
			ID:         223,
			TenantID:   tenantID,
			Code:       inviteCode,
			Email:      emails[1],
			UserID:     &users[1].Id,
			Status:     enum.Sent,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
			Edges:      ent.InviteEdges{},
		},
	}

	expectedInvites := service.ReplyInvites(
		[]biz.InviteItem{
			{
				Invite: entInvites[0],
				User:   users[0],
			},
			{
				Invite: entInvites[1],
				User:   users[1],
			},
		},
	)

	iamRemote.EXPECT().GetUsers(
		gomock.Any(), &iam_v1.GetUsersRequest{Emails: emails},
	).Return(&iam_v1.UsersReply{Users: users}, nil)
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, inviteDtos).Return(entInvites, nil)
	tenantsRepo.EXPECT().GetTenant(gomock.Any(), tenantID).Return(
		&ent.Tenant{}, nil,
	).Do(func() { time.Sleep(time.Second) }).AnyTimes() // this method is used in goroutine, so we need to wait

	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)

	invites, err := invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     invitesDto.RoleID,
			Resource:   invitesDto.Resource,
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.NoError(t, err)

	require.Equal(t, expectedInvites, invites.GetInvites())
}

func TestFailVerify(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	queue := nats_mock.NewMockIQueueManager(ctrl)
	invitesRepo := mock.NewMockInvitesRepo(ctrl)
	tenantsRepo := mock.NewMockTenantsRepo(ctrl)
	iamRemote := mock.NewMockIIamRemote(ctrl)
	rbacRemote := mock.NewMockIRbacRemote(ctrl)
	config := mock.NewMockConfig(ctrl)

	invitesUsecase, err := biz.NewInvitesUsecase(
		logger, tenantsRepo, invitesRepo, iamRemote, rbacRemote, queue, config,
	)
	require.NoError(t, err)

	invitesService := service.NewInvitesService(nil, invitesUsecase)

	ctx := context.Background()

	var tenantID int64 = 12
	var actorID int64 = 332
	appID := "pms"
	emails := []string{"email1", "email2"}
	invitesDto := &data.InvitesDTO{
		Emails:     emails,
		Lang:       "en",
		RoleID:     5, // project viewer
		Resource:   "project",
		ResourceID: 223,
	}

	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)

	_, err = invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     invitesDto.RoleID,
			Resource:   "",
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.Error(t, err)

	_, err = invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     invitesDto.RoleID,
			Resource:   invitesDto.Resource,
			ResourceId: 0,
		},
	)
	require.Error(t, err)

	_, err = invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     0,
			Resource:   invitesDto.Resource,
			ResourceId: 0,
		},
	)
	require.Error(t, err)

	_, err = invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     0,
			Resource:   "",
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.Error(t, err)

	_, err = invitesService.CreateInvites(
		ctx, &v1.CreateInvitesRequest{
			Emails:     invitesDto.Emails,
			AppId:      appID,
			Language:   invitesDto.Lang,
			RoleId:     0,
			Resource:   invitesDto.Resource,
			ResourceId: invitesDto.ResourceID,
		},
	)
	require.Error(t, err)
}
