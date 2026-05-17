package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type TenantsService struct {
	v1.UnimplementedTenantsServer

	tu *biz.TenantsUsecase
	mu *biz.MembersUsecase
}

func NewTenantsService(
	tu *biz.TenantsUsecase,
	mu *biz.MembersUsecase,
) *TenantsService {
	return &TenantsService{
		tu: tu,
		mu: mu,
	}
}

func (s *TenantsService) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.TenantReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	dto := data.TenantDto{
		OwnerID: actorID,
		Name:    req.GetName(),
		Type:    enum.TenantType(req.GetType()).DefaultIfInvalid(),
	}
	if req.ReferredBy != nil {
		rb := *req.ReferredBy
		dto.ReferredBy = &rb
	}

	tenant, err := s.tu.CreateTenant(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) UpdateTenant(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.UpdateTenant(
		ctx, data.TenantDto{
			TenantID: req.GetTenantId(),
			Name:     req.GetName(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) DeleteTenant(ctx context.Context, req *v1.TenantRequest) (*utils_v1.EmptyReply, error) {
	err := s.tu.DeleteTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *TenantsService) DeleteUsersTenants(
	ctx context.Context,
	req *v1.DeleteUsersTenantsRequest,
) (*utils_v1.EmptyReply, error) {
	err := s.tu.DeleteUsersTenants(ctx, req.GetUsersIds())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *TenantsService) GetTenant(ctx context.Context, req *v1.TenantRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.GetTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) ListTenants(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)

	list, err := s.tu.ListTenants(
		ctx, data.TenantsListFilter{
			UserID:  actorID,
			OwnerID: req.GetOwnerId(),
		}, req.GetPaginate(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListTenantsReply{
		Tenants:  replyTenants(list.Tenants),
		Paginate: list.Paginate,
	}, nil
}

func (s *TenantsService) TransferOwnership(ctx context.Context, req *v1.TransferOwnershipRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.TransferOwnership(ctx, req.GetTenantId(), req.GetNewOwnerId())
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) GetReferredTenants(ctx context.Context, req *v1.GetReferredTenantsRequest) (*v1.ListTenantsReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	list, err := s.tu.GetReferredTenants(ctx, actorID, req.GetPaginate())
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
		Type:      tenant.Type.Value(),
	}
	if tenant.ReferredBy != nil {
		result.ReferredBy = tenant.ReferredBy
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
