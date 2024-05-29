package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type MembersService struct {
	v1.UnimplementedMembersServer

	tu *biz.TenantsUsecase
	mu *biz.MembersUsecase
}

func NewMembersService(
	tu *biz.TenantsUsecase,
	mu *biz.MembersUsecase,
) *MembersService {
	return &MembersService{
		tu: tu,
		mu: mu,
	}
}

func (s *MembersService) GetMembersByIdentities(ctx context.Context, req *v1.IdentitiesRequest) (*v1.MembersReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	members, err := s.mu.GetMembersByIdentities(ctx, tenantId, req.IdentityIds)
	if err != nil {
		return nil, err
	}

	return &v1.MembersReply{
		Members: replyMembers(members),
	}, nil
}

func (s *MembersService) CreateMembers(ctx context.Context, req *v1.CreateMembersRequest) (*utils_v1.EmptyReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	tenant, err := s.tu.GetTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	// TODO: nobody can add members directly
	if tenant.OwnerID != actorId {
		return nil, v1.ErrorForbidden("only owner can add members")
	}

	_, err = s.mu.CreateMembers(ctx, tenantId, req.UsersIds)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) DeleteMember(ctx context.Context, req *v1.MemberRequest) (*utils_v1.EmptyReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	err := s.mu.DeleteMember(ctx, tenantId, req.MemberId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.MemberRequest) (*v1.MemberReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	member, err := s.mu.GetMember(ctx, tenantId, req.MemberId)
	if err != nil {
		return nil, err
	}

	return &v1.MemberReply{
		Member: replyMember(member),
	}, nil
}

func (s *MembersService) GetMemberIdentities(ctx context.Context, req *v1.GetMemberIdentitiesRequest) (*v1.GetMemberIdentitiesReply, error) {
	member, err := s.mu.GetMemberByUserId(ctx, req.TenantId, req.UserId)
	if err != nil {
		return nil, err
	}

	result := v1.GetMemberIdentitiesReply{
		Member: member.IdentityID.String(),
	}

	if len(member.Edges.Groups) > 0 {
		groups := make([]string, len(member.Edges.Groups))
		for i, group := range member.Edges.Groups {
			groups[i] = group.IdentityID.String()
		}

		result.Groups = groups
	}

	return &result, nil
}

func (s *MembersService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	filter := data.MembersListFilter{
		TenantId:       tenantId,
		GroupId:        req.GetGroupId(),
		Search:         req.GetSearch(),
		ExcludeGroupId: req.GetExcludeGroupId(),
	}

	list, err := s.mu.ListMembers(ctx, filter, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}

	return &v1.ListMembersReply{
		Members:  replyMembers(list.Members),
		Paginate: list.Paginate,
	}, nil
}

func replyMember(member *biz.MemberItem) *v1.TenantMember {
	identityId := member.IdentityID.String()
	result := v1.TenantMember{
		Id:         member.ID,
		IdentityId: &identityId,
		CreatedAt:  member.CreatedAt.Format(time.RFC3339),
		UserId:     member.UserID,
	}

	if len(member.Edges.Groups) > 0 {
		groups := make([]int64, len(member.Edges.Groups))
		for i, group := range member.Edges.Groups {
			groups[i] = group.ID
		}

		result.Groups = groups
	}

	return &result
}

func replyMembers(members []*biz.MemberItem) []*v1.TenantMember {
	reply := make([]*v1.TenantMember, len(members))
	for i, member := range members {
		reply[i] = replyMember(member)
	}
	return reply
}
