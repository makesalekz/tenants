package biz

import (
	"context"

	v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/internal/conf"
	"tenants/internal/data"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
)

type TenantsList struct {
	Tenants  []*ent.Tenant
	Paginate *v1.PaginateReply
}

// TenantsUsecase is a Greeter usecase.
type TenantsUsecase struct {
	conf      *conf.Bootstrap
	log       *log.Helper
	discovery *consul.Registry
	jwt       *data.JwtProcessor
	repo      data.TenantsRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewTenantsUsecase(logger log.Logger, c *data.Config, jwt *data.JwtProcessor, repo data.TenantsRepo) (*TenantsUsecase, error) {
	return &TenantsUsecase{
		conf:      c.Bootstrap,
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		jwt:       jwt,
		repo:      repo,
	}, nil
}

func (uc *TenantsUsecase) CreateTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	userId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	dto.OwnerId = userId

	return uc.repo.CreateTenant(ctx, dto)
}

func (uc *TenantsUsecase) UpdateTenant(ctx context.Context, teamId int64, dto data.TenantDto) (*ent.Tenant, error) {
	userId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	dto.OwnerId = userId

	return uc.repo.UpdateTenant(ctx, teamId, dto)
}

func (uc *TenantsUsecase) DeleteTenant(ctx context.Context, teamId int64) error {
	userId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return v1.ErrorUnauthorized("Unauthorized")
	}

	return uc.repo.DeleteTenant(ctx, teamId, userId)
}

func (uc *TenantsUsecase) GetTenant(ctx context.Context, teamId int64) (*ent.Tenant, error) {
	return uc.repo.GetTenant(ctx, teamId)
}

func (uc *TenantsUsecase) ListTenants(ctx context.Context, filter data.TenantsListFilter, paginate *v1.PaginateRequest) (*TenantsList, error) {
	if paginate == nil {
		paginate = &v1.PaginateRequest{}
	}

	tenants, err := uc.repo.ListTenants(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.CountListTenants(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := v1.PaginateReply{
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
