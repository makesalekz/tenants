package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
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

func (s *MembersService) CreateMembers(ctx context.Context, req *v1.CreateMembersRequest) (*utils_v1.EmptyReply, error) {
	_, err := s.mu.CreateMembers(ctx, req.UsersIds)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) DeleteMembers(ctx context.Context, req *v1.DeleteMemberRequest) (*utils_v1.EmptyReply, error) {
	err := s.mu.DeleteMember(ctx, req.MemberId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *MembersService) GetMember(ctx context.Context, req *v1.GetMemberRequest) (*v1.GetMemberReply, error) {
	member, err := s.mu.GetMember(ctx, req.UserId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("Member not found")
		}
		return nil, err
	}
	return &v1.GetMemberReply{
		Member: member.IdentityID.String(),
		Groups: []string{},
	}, nil
}

func (s *MembersService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersReply, error) {
	list, err := s.mu.ListMembers(ctx, req.Search, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}
	return &v1.ListMembersReply{
		Members:  replyMembers(list.Members),
		Paginate: list.Paginate,
	}, nil
}
