package service

import (
	"context"
	"time"

	v1 "github.com/makesalekz/tenants/api/tenants/v1"
	"github.com/makesalekz/tenants/ent"
	"github.com/makesalekz/tenants/internal/biz"
	"github.com/makesalekz/tenants/internal/data"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"
	"github.com/makesalekz/utils/v2/auth"
)

type GroupsService struct {
	v1.UnimplementedGroupsServer

	tu *biz.TenantsUsecase
	mu *biz.GroupsUsecase
}

func NewGroupsService(
	tu *biz.TenantsUsecase,
	mu *biz.GroupsUsecase,
) *GroupsService {
	return &GroupsService{
		tu: tu,
		mu: mu,
	}
}

func (s *GroupsService) CreateGroup(ctx context.Context, req *v1.CreateGroupRequest) (*v1.GroupReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.CreateGroup(
		ctx, actorID, data.CreateGroupDto{
			TenantID:    tenantID,
			Name:        req.GetName(),
			Description: req.GetDescription(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(group),
	}, nil
}

func (s *GroupsService) UpdateGroup(ctx context.Context, req *v1.UpdateGroupRequest) (*v1.GroupReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.GetGroup(ctx, tenantID, req.GetGroupId())
	if err != nil {
		return nil, err
	}

	updated, err := s.mu.UpdateGroup(
		ctx, group, data.UpdateGroupDto{
			Name:        req.GetName(),
			Description: req.GetDescription(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(updated),
	}, nil
}

func (s *GroupsService) DeleteGroup(ctx context.Context, req *v1.GroupRequest) (*utils_v1.EmptyReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.GetGroup(ctx, tenantID, req.GetGroupId())
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
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.GetGroup(ctx, tenantID, req.GetGroupId())
	if err != nil {
		return nil, err
	}

	return &v1.GroupReply{
		Group: groupReply(group),
	}, nil
}

func (s *GroupsService) ListGroups(ctx context.Context, req *v1.ListGroupsRequest) (*v1.ListGroupsReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	filter := data.GroupsListFilter{
		TenantID: tenantID,
		Search:   req.GetSearch(),
	}

	list, err := s.mu.ListGroups(ctx, filter, req.GetSort(), req.GetPaginate())
	if err != nil {
		return nil, err
	}

	return &v1.ListGroupsReply{
		Groups:   groupsReply(list.Groups),
		Paginate: list.Paginate,
	}, nil
}

func (s *GroupsService) AddMembersToGroup(ctx context.Context, req *v1.GroupMembersRequest) (
	*utils_v1.EmptyReply, error,
) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.GetGroup(ctx, tenantID, req.GetGroupId())
	if err != nil {
		return nil, err
	}

	err = s.mu.AddMembersToGroup(ctx, group, req.GetMembersIds())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *GroupsService) RemoveMembersFromGroup(ctx context.Context, req *v1.GroupMembersRequest) (
	*utils_v1.EmptyReply, error,
) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	group, err := s.mu.GetGroup(ctx, tenantID, req.GetGroupId())
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
		IdentityId:  group.IdentityID.String(),
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
