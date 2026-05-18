package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	iam_v1 "github.com/makesalekz/iam/api/iam/v1"
	v1 "github.com/makesalekz/tenants/api/tenants/v1"
	"github.com/makesalekz/tenants/ent"
	"github.com/makesalekz/tenants/ent/enum"
	"github.com/makesalekz/tenants/internal/biz"
	"github.com/makesalekz/tenants/internal/data"
	"github.com/makesalekz/tenants/internal/data/mock"
	"github.com/makesalekz/tenants/internal/service"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"
	"github.com/makesalekz/utils/v2/auth"
	"github.com/makesalekz/utils/v2/zap"
)

func beforeTenantsTest(t *testing.T) (
	context.Context,
	*service.TenantsService,
	*gomock.Controller,
	*mock.MockTenantsRepo,
	*mock.MockIIamRemote,
	int64,
	int64,
) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)
	tenantsRepo := mock.NewMockTenantsRepo(ctrl)
	membersRepo := mock.NewMockMembersRepo(ctrl)
	iamRemote := mock.NewMockIIamRemote(ctrl)
	rbacRemote := mock.NewMockIRbacRemote(ctrl)

	tenantsUsecase, err := biz.NewTenantsUsecase(logger, tenantsRepo, rbacRemote, iamRemote)
	require.NoError(t, err)

	membersUsecase, err := biz.NewMembersUsecase(tenantsRepo, membersRepo, iamRemote)
	require.NoError(t, err)

	tenantsService := service.NewTenantsService(tenantsUsecase, membersUsecase)

	var tenantID int64 = 12
	var actorID int64 = 332
	ctx := auth.NewTenantContext(auth.NewActorContext(context.Background(), actorID), tenantID)

	return ctx, tenantsService, ctrl, tenantsRepo, iamRemote, tenantID, actorID
}

func TestTransferOwnership(t *testing.T) {
	ctx, svc, ctrl, tenantsRepo, iamRemote, _, _ := beforeTenantsTest(t)
	defer ctrl.Finish()

	var newOwnerID int64 = 500
	var tenantID int64 = 12
	var referredBy int64 = 332

	iamRemote.EXPECT().GetUser(gomock.Any(), newOwnerID).Return(&iam_v1.UserShort{
		Id: newOwnerID,
	}, nil)

	tenantsRepo.EXPECT().TransferOwnership(gomock.Any(), tenantID, newOwnerID).Return(
		&ent.Tenant{
			ID:         tenantID,
			OwnerID:    newOwnerID,
			Name:       "Test Tenant",
			ReferredBy: &referredBy,
			CreatedAt:  time.Now(),
			Type:       enum.Business,
		}, nil,
	)

	reply, err := svc.TransferOwnership(ctx, &v1.TransferOwnershipRequest{
		TenantId:   tenantID,
		NewOwnerId: newOwnerID,
	})

	require.NoError(t, err)
	require.Equal(t, newOwnerID, reply.Tenant.OwnerId)
	// referred_by preserved
	require.NotNil(t, reply.Tenant.ReferredBy)
	require.Equal(t, referredBy, *reply.Tenant.ReferredBy)
}

func TestTransferOwnership_NewOwnerNotFound(t *testing.T) {
	ctx, svc, ctrl, _, iamRemote, _, _ := beforeTenantsTest(t)
	defer ctrl.Finish()

	var newOwnerID int64 = 999

	iamRemote.EXPECT().GetUser(gomock.Any(), newOwnerID).Return(nil, &ent.NotFoundError{})

	_, err := svc.TransferOwnership(ctx, &v1.TransferOwnershipRequest{
		TenantId:   12,
		NewOwnerId: newOwnerID,
	})

	require.Error(t, err)
}

func TestGetReferredTenants(t *testing.T) {
	ctx, svc, ctrl, tenantsRepo, _, _, actorID := beforeTenantsTest(t)
	defer ctrl.Finish()

	referredBy := actorID

	tenantsRepo.EXPECT().ListTenants(gomock.Any(), data.TenantsListFilter{
		ReferredBy: actorID,
	}, gomock.Any()).Return(
		[]*ent.Tenant{
			{
				ID:         100,
				OwnerID:    500,
				Name:       "Shop A",
				ReferredBy: &referredBy,
				CreatedAt:  time.Now(),
				Type:       enum.Business,
			},
			{
				ID:         101,
				OwnerID:    501,
				Name:       "Shop B",
				ReferredBy: &referredBy,
				CreatedAt:  time.Now(),
				Type:       enum.Business,
			},
		}, nil,
	)

	tenantsRepo.EXPECT().CountListTenants(gomock.Any(), data.TenantsListFilter{
		ReferredBy: actorID,
	}).Return(int32(2), nil)

	reply, err := svc.GetReferredTenants(ctx, &v1.GetReferredTenantsRequest{
		Paginate: &utils_v1.PaginateRequest{Limit: 10},
	})

	require.NoError(t, err)
	require.Len(t, reply.Tenants, 2)
	require.Equal(t, "Shop A", reply.Tenants[0].Name)
	require.Equal(t, "Shop B", reply.Tenants[1].Name)
}
