package service

import (
	"context"
	"time"

	v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/internal/biz"
	"tenants/internal/data"
)

type MembersService struct {
	v1.UnimplementedMembersServer

	mu *biz.MembersUsecase
}

func NewMembersService(mu *biz.MembersUsecase) *MembersService {
	return &MembersService{
		mu: mu,
	}
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

func (s *MembersService) CreateMembers(ctx context.Context, req *v1.CreateMembersRequest) (*v1.CreateMembersReply, error) {
	_, err := s.mu.CreateMembers(ctx, req.TenantId, req.UsersIds)
	if err != nil {
		return nil, err
	}

	return &v1.CreateMembersReply{}, nil
}

func (s *MembersService) DeleteMembers(ctx context.Context, req *v1.DeleteMemberRequest) (*v1.DeleteMemberReply, error) {
	err := s.mu.DeleteMember(ctx, req.TenantId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteMemberReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.GetMemberRequest) (*v1.GetMemberReply, error) {
	member, err := s.mu.GetMember(ctx, req.TenantId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.GetMemberReply{
		Member: member.IdentityID.String(),
	}, nil
}

func (s *MembersService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersReply, error) {
	list, err := s.mu.ListMembers(ctx, data.MembersListFilter{
		TenantId: req.TenantId,
	}, req.Paginate)
	if err != nil {
		return nil, err
	}
	return &v1.ListMembersReply{
		Members:  replyMembers(list.Members),
		Paginate: list.Paginate,
	}, nil
}
