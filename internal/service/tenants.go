package service

import (
	"context"
	"time"

	v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/internal/biz"
	"tenants/internal/data"
)

type TenantsService struct {
	v1.UnimplementedTenantsServer

	tu *biz.TenantsUsecase
	mu *biz.MembersUsecase
}

func NewTenantsService(tu *biz.TenantsUsecase, mu *biz.MembersUsecase) *TenantsService {
	return &TenantsService{
		tu: tu,
		mu: mu,
	}
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

func (s *TenantsService) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.CreateTenant(ctx, data.TenantDto{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) UpdateCurrentTenant(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.UpdateCurrentTenant(ctx, data.TenantDto{
		Name: req.Name,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) DeleteCurrentTenant(ctx context.Context, req *v1.EmptyRequest) (*v1.EmptyReply, error) {
	err := s.tu.DeleteCurrentTenant(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *TenantsService) GetCurrentTenant(ctx context.Context, req *v1.EmptyRequest) (*v1.TenantReply, error) {
	tenant, err := s.tu.GetCurrentTenant(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *TenantsService) ListTenants(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsReply, error) {
	list, err := s.tu.ListTenants(ctx, data.TenantsListFilter{
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
