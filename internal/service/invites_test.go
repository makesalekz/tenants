package service_test

import (
	"context"
	"errors"
	"fmt"
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
	"gitlab.calendaria.team/services/utils/v2/auth"
	nats_mock "gitlab.calendaria.team/services/utils/v2/nats/mock"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"
	"gitlab.calendaria.team/services/utils/v2/zap"
)

func beforeTest(t *testing.T) (
	context.Context,
	*service.InvitesService,
	*gomock.Controller,
	*mock.MockInvitesRepo,
	*mock.MockIIamRemote,
	int64,
	int64,
) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)
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
	iamRemote.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&iam_v1.UserShort{}, nil).AnyTimes()
	ctx := context.Background()
	var tenantID int64 = 12
	var actorID int64 = 332
	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)

	return ctx, invitesService, ctrl, invitesRepo, iamRemote, tenantID, actorID
}

func TestInvitesCreate(t *testing.T) {
	ctx, invitesService, ctrl, invitesRepo, iamRemote, tenantID, actorID := beforeTest(t)
	defer ctrl.Finish()

	appID := ""
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
		{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     5,
			Resource:   "project",
			ResourceID: 223,
		},
		{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     5,
			Resource:   "project",
			ResourceID: 223,
		},
	}

	entInvites := []*ent.Invite{
		{
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
		{
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
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ int64, dtos []data.InviteDto) ([]*ent.Invite, error) {
			if !containsAll(inviteDtos, dtos) {
				return nil, errors.New("elements do not match")
			}
			return entInvites, nil
		},
	).Times(1)

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
	ctx, invitesService, ctrl, invitesRepo, iamRemote, tenantID, actorID := beforeTest(t)
	defer ctrl.Finish()

	appID := ""
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
		{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
		},
		{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     5,
			Resource:   "",
			ResourceID: 0,
		},
	}

	entInvites := []*ent.Invite{
		{
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
		{
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
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ int64, dtos []data.InviteDto) ([]*ent.Invite, error) {
			// Проверяем, что все элементы присутствуют с использованием вспомогательной функции
			if !containsAll(inviteDtos, dtos) {
				return nil, errors.New("elements do not match")
			}
			return entInvites, nil
		},
	).Times(1)

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
	ctx, invitesService, ctrl, invitesRepo, iamRemote, tenantID, actorID := beforeTest(t)
	defer ctrl.Finish()

	appID := ""
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
		{
			Email:      emails[0],
			UserID:     &users[0].Id,
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
		},
		{
			Email:      emails[1],
			UserID:     &users[1].Id,
			RoleID:     0,
			Resource:   "",
			ResourceID: 0,
		},
	}

	entInvites := []*ent.Invite{
		{
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
		{
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
	invitesRepo.EXPECT().CreateInvites(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ int64, dtos []data.InviteDto) ([]*ent.Invite, error) {
			if !containsAll(inviteDtos, dtos) {
				return nil, errors.New("elements do not match")
			}
			return entInvites, nil
		},
	).Times(1)

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
	ctx, invitesService, ctrl, _, _, tenantID, actorID := beforeTest(t)

	defer ctrl.Finish()

	appID := ""
	emails := []string{"email1", "email2"}
	invitesDto := &data.InvitesDTO{
		Emails:     emails,
		Lang:       "en",
		RoleID:     5, // project viewer
		Resource:   "project",
		ResourceID: 223,
	}

	ctx = auth.NewTenantContext(auth.NewActorContext(ctx, actorID), tenantID)
	var err error

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

func containsAll(expected, actual []data.InviteDto) bool {
	if len(expected) != len(actual) {
		return false
	}

	expectedMap := make(map[string]int)
	for _, item := range expected {
		key := fmt.Sprintf("%s-%d-%s-%d", item.Email, item.RoleID, item.Resource, item.ResourceID)
		expectedMap[key]++
	}

	for _, item := range actual {
		key := fmt.Sprintf("%s-%d-%s-%d", item.Email, item.RoleID, item.Resource, item.ResourceID)
		if count, ok := expectedMap[key]; !ok || count == 0 {
			return false
		}
		expectedMap[key]--
	}

	for _, count := range expectedMap {
		if count != 0 {
			return false
		}
	}

	return true
}
