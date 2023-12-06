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

type GroupsService struct {
	v1.UnimplementedGroupsServer

	jwt *jwt.JwtProcessor
	tu  *biz.TenantsUsecase
	mu  *biz.GroupsUsecase
}

func NewGroupsService(jwt *jwt.JwtProcessor, tu *biz.TenantsUsecase, mu *biz.GroupsUsecase) *GroupsService {
	return &GroupsService{
		jwt: jwt,
		tu:  tu,
		mu:  mu,
	}
}

func (s *GroupsService) CreateGroup(ctx context.Context, req *v1.CreateGroupRequest) (*v1.GroupReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}
	// TODO: check permissions

	group, err := s.mu.CreateGroup(ctx, data.CreateGroupDto{
		TenantId:    claims.GetTenantId(),
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(group),
	}, nil
}

func (s *GroupsService) UpdateGroup(ctx context.Context, req *v1.UpdateGroupRequest) (*v1.GroupReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	// TODO: check permissions

	group, err := s.mu.GetGroup(ctx, claims.GetTenantId(), req.GetGroupId())
	if err != nil {
		return nil, err
	}

	updated, err := s.mu.UpdateGroup(ctx, group, data.UpdateGroupDto{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(updated),
	}, nil
}

func (s *GroupsService) DeleteGroup(ctx context.Context, req *v1.GroupRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}
	// TODO: check permissions

	group, err := s.mu.GetGroup(ctx, claims.GetTenantId(), req.GetGroupId())
	if err != nil {
		return nil, err
	}

	err = s.mu.DeleteGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *GroupsService) GetGroup(ctx context.Context, req *v1.GroupRequest) (*v1.GroupReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}
	// TODO: check permissions

	group, err := s.mu.GetGroup(ctx, claims.GetTenantId(), req.GetGroupId())
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(group),
	}, nil
}

func (s *GroupsService) ListGroups(ctx context.Context, req *v1.ListGroupsRequest) (*v1.ListGroupsReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	filter := data.GroupsListFilter{
		TenantId: claims.GetTenantId(),
		Search:   req.Search,
	}

	list, err := s.mu.ListGroups(ctx, filter, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}

	return &v1.ListGroupsReply{
		Groups:   groupsReply(list.Groups),
		Paginate: list.Paginate,
	}, nil
}

func (s *GroupsService) AddMembersToGroup(ctx context.Context, req *v1.GroupMembersRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}
	// TODO: check permissions

	group, err := s.mu.GetGroup(ctx, claims.GetTenantId(), req.GetGroupId())
	if err != nil {
		return nil, err
	}

	err = s.mu.AddMembersToGroup(ctx, group, req.GetMembersIds())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *GroupsService) RemoveMembersFromGroup(ctx context.Context, req *v1.GroupMembersRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}
	// TODO: check permissions

	group, err := s.mu.GetGroup(ctx, claims.GetTenantId(), req.GetGroupId())
	if err != nil {
		return nil, err
	}

	err = s.mu.RemoveMembersFromGroup(ctx, group, req.GetMembersIds())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func groupReply(group *ent.Group) *v1.Group {
	return &v1.Group{
		Id:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
	}
}

func groupsReply(groups []*ent.Group) []*v1.Group {
	reply := make([]*v1.Group, len(groups))
	for i, group := range groups {
		reply[i] = groupReply(group)
	}
	return reply
}
