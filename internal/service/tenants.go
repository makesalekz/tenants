package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type TenantsService struct {
	v1.UnimplementedTenantsServer

	jwt *jwt.JwtProcessor
	tu  *biz.TenantsUsecase
	mu  *biz.MembersUsecase
}

func NewTenantsService(jwt *jwt.JwtProcessor, tu *biz.TenantsUsecase, mu *biz.MembersUsecase) *TenantsService {
	return &TenantsService{
		jwt: jwt,
		tu:  tu,
		mu:  mu,
	}
}

func (s *TenantsService) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.TenantReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := s.tu.CreateTenant(ctx, data.TenantDto{
		OwnerId: claims.GetUserId(),
		Name:    req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) UpdateCurrentTenant(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.TenantReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions
	tenant, err := s.tu.UpdateTenant(ctx, data.TenantDto{
		TenantId: claims.GetTenantId(),
		OwnerId:  claims.GetUserId(),
		Name:     req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) DeleteCurrentTenant(ctx context.Context, req *utils_v1.EmptyRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions
	err := s.tu.DeleteTenant(ctx, data.TenantDto{
		TenantId: claims.GetTenantId(),
		OwnerId:  claims.GetUserId(),
	})
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *TenantsService) GetCurrentTenant(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.TenantReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions
	tenant, err := s.tu.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) ListTenants(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsReply, error) {
	// TODO: check permissions to get all tenants
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	list, err := s.tu.ListTenants(ctx, data.TenantsListFilter{
		UserId:  claims.GetUserId(),
		OwnerId: req.OwnerId,
	}, req.Paginate)
	if err != nil {
		return nil, err
	}
	return &v1.ListTenantsReply{
		Tenants:  replyTenants(list.Tenants),
		Paginate: list.Paginate,
	}, nil
}

func replyTenant(tenant *ent.Tenant) *v1.Tenant {
	result := v1.Tenant{
		Id:        tenant.ID,
		OwnerId:   tenant.OwnerID,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt.Format(time.RFC3339),
	}

	return &result
}

func replyTenants(tenants []*ent.Tenant) []*v1.Tenant {
	result := make([]*v1.Tenant, len(tenants))
	for i, tenant := range tenants {
		result[i] = replyTenant(tenant)
	}
	return result
}
