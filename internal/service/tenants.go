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

func replyTenant(team *ent.Tenant) *v1.Tenant {
	result := v1.Tenant{
		Id:        team.ID,
		OwnerId:   team.OwnerID,
		Name:      team.Name,
		CreatedAt: team.CreatedAt.Format(time.RFC3339),
	}

	return &result
}

func replyTenants(tenants []*ent.Tenant) []*v1.Tenant {
	result := make([]*v1.Tenant, len(tenants))
	for i, team := range tenants {
		result[i] = replyTenant(team)
	}
	return result
}

func (s *TenantsService) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.TenantReply, error) {
	team, err := s.tu.CreateTenant(ctx, data.TenantDto{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(team),
	}, nil
}

func (s *TenantsService) UpdateTenant(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.TenantReply, error) {
	team, err := s.tu.UpdateTenant(ctx, req.TenantId, data.TenantDto{
		Name: req.Name,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}
	return &v1.TenantReply{
		Tenant: replyTenant(team),
	}, nil
}

func (s *TenantsService) DeleteTenant(ctx context.Context, req *v1.TenantRequest) (*v1.EmptyReply, error) {
	err := s.tu.DeleteTenant(ctx, req.TenantId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *TenantsService) GetTenant(ctx context.Context, req *v1.TenantRequest) (*v1.TenantReply, error) {
	team, err := s.tu.GetTenant(ctx, req.TenantId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Tenant not found")
		}
		return nil, err
	}

	result := &v1.TenantReply{
		Tenant: replyTenant(team),
	}

	member, _ := s.mu.GetOwnMember(ctx, team.ID)
	if member != nil {
		result.MemberId = &member.ID
	}

	return result, nil
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
