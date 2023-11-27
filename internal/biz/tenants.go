package biz

import (
	"context"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type TenantsList struct {
	Tenants  []*ent.Tenant
	Paginate *utils_v1.PaginateReply
}

// TenantsUsecase is a Greeter usecase.
type TenantsUsecase struct {
	log       *log.Helper
	discovery *consul.Registry
	jwt       *jwt.JwtProcessor
	repo      data.TenantsRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewTenantsUsecase(logger log.Logger, c *config.Config, jwt *jwt.JwtProcessor, repo data.TenantsRepo) (*TenantsUsecase, error) {
	return &TenantsUsecase{
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		jwt:       jwt,
		repo:      repo,
	}, nil
}

func (uc *TenantsUsecase) CreateTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	return uc.repo.CreateTenant(ctx, dto)
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

func (uc *TenantsUsecase) DeleteTenant(ctx context.Context, dto data.TenantDto) error {
	err := uc.repo.DeleteTenant(ctx, dto)
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
