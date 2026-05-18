package biz

import (
	"context"
	"strconv"

	rbac_v1 "github.com/makesalekz/rbac/api/rbac/v1"
	tenants_v1 "github.com/makesalekz/tenants/api/tenants/v1"
	"github.com/makesalekz/tenants/ent"
	"github.com/makesalekz/tenants/internal/data"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
)

const (
	AdminRoleID        = 1
	BasicRoleID        = 2
	QalaiBasicRoleID   = 7
	QalaiTrialRoleID   = 8
	BasQaruBasicRoleID = 9
)

type TenantsList struct {
	Tenants  []*ent.Tenant
	Paginate *utils_v1.PaginateReply
}

// TenantsUsecase is a Greeter usecase.
type TenantsUsecase struct {
	log *log.Helper

	repo data.TenantsRepo
	rbac data.IRbacRemote
	iam  data.IIamRemote
}

// NewGreeterUsecase new a Greeter usecase.
func NewTenantsUsecase(
	logger log.Logger,
	repo data.TenantsRepo,
	rbac data.IRbacRemote,
	iam data.IIamRemote,
) (*TenantsUsecase, error) {
	return &TenantsUsecase{
		log:  log.NewHelper(logger),
		repo: repo,
		rbac: rbac,
		iam:  iam,
	}, nil
}

func (uc *TenantsUsecase) CreateTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	tenant, member, err := uc.repo.CreateTenant(ctx, dto)
	if err != nil {
		return nil, err
	}

	err = uc.rbac.AssignRoles(
		metadata.AppendToClientContext(ctx, "x-md-global-tenant-id", strconv.FormatInt(tenant.ID, 10)),
		&rbac_v1.AssignRoleRequest{
			IdentityId: member.IdentityID.String(),
			RoleId:     AdminRoleID,
		},
		&rbac_v1.AssignRoleRequest{
			IdentityId: "",
			RoleId:     BasicRoleID,
		},
		&rbac_v1.AssignRoleRequest{
			IdentityId: "",
			RoleId:     QalaiBasicRoleID,
		},
		&rbac_v1.AssignRoleRequest{
			IdentityId: "",
			RoleId:     QalaiTrialRoleID,
		},
		&rbac_v1.AssignRoleRequest{
			IdentityId: "",
			RoleId:     BasQaruBasicRoleID,
		},
	)
	if err != nil {
		uc.log.Errorf("CreateTenant.AssignRoles: %s", err.Error())
	}

	return tenant, nil
}

func (uc *TenantsUsecase) UpdateTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	tenant, err := uc.repo.UpdateTenant(ctx, dto)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}
	return tenant, nil
}

func (uc *TenantsUsecase) DeleteTenant(ctx context.Context, tenantID int64) error {
	err := uc.repo.DeleteTenant(ctx, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return tenants_v1.ErrorNotFound("tenant not found")
		}
		return err
	}
	return nil
}

func (uc *TenantsUsecase) DeleteUsersTenants(ctx context.Context, usersIDs []int64) error {
	_, err := uc.repo.DeleteUsersTenants(ctx, usersIDs)
	if err != nil {
		if ent.IsNotFound(err) {
			return tenants_v1.ErrorNotFound("tenant not found")
		}
		return err
	}

	return nil
}

func (uc *TenantsUsecase) GetTenant(ctx context.Context, tenantID int64) (*ent.Tenant, error) {
	tenant, err := uc.repo.GetTenant(ctx, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}
	return tenant, nil
}

func (uc *TenantsUsecase) ListTenants(
	ctx context.Context,
	filter data.TenantsListFilter,
	paginate *utils_v1.PaginateRequest,
) (*TenantsList, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	tenants, err := uc.repo.ListTenants(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.CountListTenants(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := utils_v1.PaginateReply{
		Total: &total,
	}

	if len(tenants) == int(paginate.GetLimit()) {
		paginateReply.FromId = &tenants[len(tenants)-1].ID
	}

	return &TenantsList{
		Tenants:  tenants,
		Paginate: &paginateReply,
	}, nil
}

func (uc *TenantsUsecase) TransferOwnership(ctx context.Context, tenantID int64, newOwnerID int64) (*ent.Tenant, error) {
	// Validate new owner exists in IAM
	_, err := uc.iam.GetUser(ctx, newOwnerID)
	if err != nil {
		return nil, tenants_v1.ErrorNotFound("new owner not found")
	}

	t, err := uc.repo.TransferOwnership(ctx, tenantID, newOwnerID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}
	return t, nil
}

func (uc *TenantsUsecase) GetReferredTenants(
	ctx context.Context,
	managerID int64,
	paginate *utils_v1.PaginateRequest,
) (*TenantsList, error) {
	filter := data.TenantsListFilter{
		ReferredBy: managerID,
	}
	return uc.ListTenants(ctx, filter, paginate)
}
