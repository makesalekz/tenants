package biz

import (
	"context"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	"gitlab.calendaria.team/services/tenants/internal/data"

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

func (uc *TenantsUsecase) UpdateCurrentTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	ownerId, claims, ok := uc.jwt.GetTenantClaimsFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	// TODO: check permissions
	dto.OwnerId = ownerId

	return uc.repo.UpdateTenant(ctx, claims.TenantId, dto)
}

func (uc *TenantsUsecase) DeleteCurrentTenant(ctx context.Context) error {
	ownerId, claims, ok := uc.jwt.GetTenantClaimsFromContext(ctx)
	if !ok {
		return v1.ErrorUnauthorized("Unauthorized")
	}

	// TODO: check permissions

	return uc.repo.DeleteTenant(ctx, claims.TenantId, ownerId)
}

func (uc *TenantsUsecase) GetCurrentTenant(ctx context.Context) (*ent.Tenant, error) {
	_, claims, ok := uc.jwt.GetTenantClaimsFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	// TODO: check permissions

	return uc.repo.GetTenant(ctx, claims.TenantId)
}

func (uc *TenantsUsecase) ListTenants(ctx context.Context, filter data.TenantsListFilter, paginate *v1.PaginateRequest) (*TenantsList, error) {
	if paginate == nil {
		paginate = &v1.PaginateRequest{}
	}

	// TODO: check permissions to get all tenants
	userId, _, ok := uc.jwt.GetTenantClaimsFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	filter.UserId = userId

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
