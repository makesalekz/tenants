package biz

import (
	"context"
	"strconv"

	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
)

const ADMIN_ROLE_ID = 1
const BASIC_ROLE_ID = 2

type TenantsList struct {
	Tenants  []*ent.Tenant
	Paginate *utils_v1.PaginateReply
}

// TenantsUsecase is a Greeter usecase.
type TenantsUsecase struct {
	log *log.Helper

	repo data.TenantsRepo
	rbac *data.RbacRemote
}

// NewGreeterUsecase new a Greeter usecase.
func NewTenantsUsecase(
	logger log.Logger,
	repo data.TenantsRepo,
	rbac *data.RbacRemote,
) (*TenantsUsecase, error) {
	return &TenantsUsecase{
		log:  log.NewHelper(logger),
		repo: repo,
		rbac: rbac,
	}, nil
}

func (uc *TenantsUsecase) CreateTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	tenant, member, err := uc.repo.CreateTenant(ctx, dto)
	if err != nil {
		return nil, err
	}

	tenantContext := metadata.AppendToClientContext(ctx, "x-md-global-tenant-id", strconv.FormatInt(tenant.ID, 10))
	tenantContext = metadata.AppendToClientContext(tenantContext, "x-md-global-actor-role", "admin")

	err = uc.rbac.AssignRole(tenantContext, member.IdentityID.String(), ADMIN_ROLE_ID)
	if err != nil {
		uc.log.Errorf("CreateTenant.AssignRole (admin): %s", err.Error())
	}

	// err = uc.rbac.AssignRole(tenantContext, "", BASIC_ROLE_ID)
	// if err != nil {
	// 	uc.log.Errorf("CreateTenant.AssignRole (basic): %s", err.Error())
	// }

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

func (uc *TenantsUsecase) DeleteTenant(ctx context.Context, tenantId int64) error {
	err := uc.repo.DeleteTenant(ctx, tenantId)
	if err != nil {
		if ent.IsNotFound(err) {
			return tenants_v1.ErrorNotFound("tenant not found")
		}
		return err
	}
	return nil

}

func (uc *TenantsUsecase) GetTenant(ctx context.Context, tenantId int64) (*ent.Tenant, error) {
	tenant, err := uc.repo.GetTenant(ctx, tenantId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}
	return tenant, nil
}

func (uc *TenantsUsecase) ListTenants(ctx context.Context, filter data.TenantsListFilter, paginate *utils_v1.PaginateRequest) (*TenantsList, error) {
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

	if len(tenants) == int(paginate.Limit) {
		paginateReply.FromId = &tenants[len(tenants)-1].ID
	}

	return &TenantsList{
		Tenants:  tenants,
		Paginate: &paginateReply,
	}, nil
}
