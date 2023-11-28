package service

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func (s *MembersService) DeleteMembers(ctx context.Context, req *v1.DeleteMemberRequest) (*utils_v1.EmptyReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	memberUUID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid member id")
	}

	tenant, err := s.tu.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can remove members")
	}

	err = s.mu.DeleteMember(ctx, claims.GetTenantId(), memberUUID)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.GetMemberRequest) (*v1.GetMemberReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	member, err := s.mu.GetMember(ctx, claims.GetTenantId(), req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.GetMemberReply{
		Member: member.IdentityID.String(),
		Groups: []string{},
	}, nil
}

func (s *MembersService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	filter := data.MembersListFilter{
		TenantId: claims.GetTenantId(),
		Search:   req.Search,
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

func replyMember(member biz.MemberItem) *v1.Member {
	return &v1.Member{
		Id:        member.ID,
		CreatedAt: member.CreatedAt.Format(time.RFC3339),
		User:      member.User,
	}
}

func replyMembers(members []biz.MemberItem) []*v1.Member {
	reply := make([]*v1.Member, len(members))
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
