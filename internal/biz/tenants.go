package biz

import (
	"context"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
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
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	dto.OwnerId = claims.GetUserId()

	return uc.repo.CreateTenant(ctx, dto)
}

func (uc *TenantsUsecase) UpdateCurrentTenant(ctx context.Context, dto data.TenantDto) (*ent.Tenant, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions
	dto.OwnerId = claims.GetUserId()

	return uc.repo.UpdateTenant(ctx, claims.GetTenantId(), dto)
}

func (uc *TenantsUsecase) DeleteCurrentTenant(ctx context.Context) error {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions

	return uc.repo.DeleteTenant(ctx, claims.GetTenantId(), claims.GetUserId())
}

func (uc *TenantsUsecase) GetCurrentTenant(ctx context.Context) (*ent.Tenant, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions

	return uc.repo.GetTenant(ctx, claims.GetTenantId())
}

func (uc *TenantsUsecase) ListTenants(ctx context.Context, filter data.TenantsListFilter, paginate *utils_v1.PaginateRequest) (*TenantsList, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	// TODO: check permissions to get all tenants
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	filter.UserId = claims.GetUserId()

	tenants, err := uc.repo.ListTenants(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	uc.log.Debug("ListTenants", "claims", claims, "filter", filter, "tenants", tenants)

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
