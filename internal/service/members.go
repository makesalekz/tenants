package service

import (
	"context"
	"time"

	"gitlab.calendaria.team/services/tenants/ent"
	"golang.org/x/exp/maps"

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

func (s *MembersService) GetShortMembers(ctx context.Context, req *v1.IdentitiesRequest) (*v1.MembersReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	reply := &v1.MembersReply{}

	members, err := s.mu.GetShortMembers(ctx, tenantID, req.GetIdentityIds(), req.GetWithGroups())
	if err != nil {
		return nil, err
	}

	reply.Members = replyShortMembers(members)

	if req.GetWithGroups() {
		groups := map[int64]*ent.Group{}
		for _, member := range members {
			for _, group := range member.Edges.Groups {
				groups[group.ID] = group
			}
		}

		reply.Groups = groupsReply(maps.Values(groups))
	}

	return reply, nil
}

func (s *MembersService) CreateMembers(ctx context.Context, req *v1.CreateMembersRequest) (
	*utils_v1.EmptyReply, error,
) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	tenant, err := s.tu.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// TODO: nobody can add members directly
	if tenant.OwnerID != actorID {
		return nil, v1.ErrorForbidden("only owner can add members")
	}

	_, err = s.mu.CreateMembers(ctx, tenantID, req.GetUsersIds())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) DeleteMember(ctx context.Context, req *v1.MemberRequest) (*utils_v1.EmptyReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	err := s.mu.DeleteMember(ctx, tenantID, req.GetMemberId())
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.MemberRequest) (*v1.MemberReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	member, err := s.mu.GetMember(ctx, tenantID, req.GetMemberId())
	if err != nil {
		return nil, err
	}

	return &v1.MemberReply{
		Id:         member.ID,
		IdentityId: member.IdentityID.String(),
		CreatedAt:  member.CreatedAt.Format(time.RFC3339),
		UserId:     member.UserID,
		Groups:     getMemberGroups(member),
	}, nil
}

func (s *MembersService) GetMemberIdentities(
	ctx context.Context, req *v1.GetMemberIdentitiesRequest,
) (*v1.GetMemberIdentitiesReply, error) {
	member, err := s.mu.GetMemberByUserID(ctx, req.GetTenantId(), req.GetUserId())
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
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	filter := data.MembersListFilter{
		TenantID:       tenantID,
		GroupID:        req.GetGroupId(),
		Search:         req.GetSearch(),
		ExcludeGroupID: req.GetExcludeGroupId(),
		WithGroups:     req.GetWithGroups(),
	}

	list, err := s.mu.ListMembers(ctx, filter, req.GetSort(), req.GetPaginate())
	if err != nil {
		return nil, err
	}

	return &v1.ListMembersReply{
		Members:  replyMembers(list.Members),
		Groups:   groupsReply(list.Groups),
		Paginate: list.Paginate,
	}, nil
}

func (s *MembersService) CountMembers(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.CountMembersReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	count, err := s.mu.CountMembers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &v1.CountMembersReply{Count: count}, nil
}

func getMemberGroups(member *biz.MemberItem) []int64 {
	if len(member.Edges.Groups) == 0 {
		return nil
	}

	groups := make([]int64, len(member.Edges.Groups))
	for i, group := range member.Edges.Groups {
		groups[i] = group.ID
	}

	return groups
}

func replyMember(member *biz.MemberItem) *v1.TenantMember {
	identityID := member.IdentityID.String()
	result := v1.TenantMember{
		Id:         member.ID,
		IdentityId: &identityID,
		CreatedAt:  member.CreatedAt.Format(time.RFC3339),
		User:       member.User,
		Groups:     getMemberGroups(member),
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

func replyShortMember(member *ent.Member) *v1.MemberReply {
	shortMember := &v1.MemberReply{
		Id:         member.ID,
		IdentityId: member.IdentityID.String(),
		CreatedAt:  member.CreatedAt.Format(time.RFC3339),
		UserId:     member.UserID,
	}

	for _, group := range member.Edges.Groups {
		shortMember.Groups = append(shortMember.Groups, group.ID)
	}

	return shortMember
}

func replyShortMembers(members []*ent.Member) []*v1.MemberReply {
	reply := make([]*v1.MemberReply, len(members))
	for i, member := range members {
		reply[i] = replyShortMember(member)
	}
	return reply
}
