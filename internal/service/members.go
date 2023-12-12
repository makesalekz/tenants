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

type MembersService struct {
	v1.UnimplementedMembersServer

	jwt *jwt.JwtProcessor
	tu  *biz.TenantsUsecase
	mu  *biz.MembersUsecase
}

func NewMembersService(jwt *jwt.JwtProcessor, tu *biz.TenantsUsecase, mu *biz.MembersUsecase) *MembersService {
	return &MembersService{
		jwt: jwt,
		tu:  tu,
		mu:  mu,
	}
}

func (s *MembersService) CreateMembers(ctx context.Context, req *v1.CreateMembersRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := s.tu.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can add members")
	}

	_, err = s.mu.CreateMembers(ctx, claims.GetTenantId(), req.UsersIds)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) DeleteMember(ctx context.Context, req *v1.MemberRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := s.tu.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can remove members")
	}

	err = s.mu.DeleteMember(ctx, claims.GetTenantId(), req.MemberId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.MemberRequest) (*v1.MemberReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	member, err := s.mu.GetMember(ctx, claims.GetTenantId(), req.MemberId)
	if err != nil {
		return nil, err
	}

	return &v1.MemberReply{
		Member: replyMember(member),
	}, nil
}

func (s *MembersService) GetMemberIdentities(ctx context.Context, req *v1.GetMemberIdentitiesRequest) (*v1.GetMemberIdentitiesReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	member, err := s.mu.GetMemberByUserId(ctx, claims.GetTenantId(), req.UserId)
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
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	filter := data.MembersListFilter{
		TenantId: claims.GetTenantId(),
		GroupId:  req.GetGroupId(),
		Search:   req.GetSearch(),
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
		User:       member.User,
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

func replyShortMembers(members []*ent.Member) []*v1.MemberShort {
	reply := make([]*v1.MemberShort, len(members))
	for i, member := range members {
		reply[i] = &v1.MemberShort{
			Id:     member.ID,
			UserId: member.UserID,
		}
	}
	return reply
}
