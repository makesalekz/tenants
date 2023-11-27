package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type InvitesService struct {
	v1.UnimplementedInvitesServer

	iu *biz.InvitesUsecase
}

func NewInvitesService(iu *biz.InvitesUsecase) *InvitesService {
	return &InvitesService{
		iu: iu,
	}
}

func (s *InvitesService) CreateInvites(ctx context.Context, req *v1.CreateInvitesRequest) (*v1.ListInvitesReply, error) {
	if len(req.Emails) == 0 {
		return nil, v1.ErrorInvalidRequest("emails is empty")
	}

	invites, err := s.iu.CreateInvites(ctx, req.Emails)
	if err != nil {
		return nil, err
	}

	return &v1.ListInvitesReply{
		Invites: replyInvites(invites),
	}, nil
}

func (s *InvitesService) UpdateInvite(ctx context.Context, req *v1.UpdateInviteRequest) (*utils_v1.EmptyReply, error) {
	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

	status := enum.InviteStatus(req.Status.String())
	if !status.IsValid() {
		return nil, v1.ErrorInvalidRequest("invalid status")
	}

	_, err := s.iu.UpdateInvite(ctx, req.InviteId, status)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) DeleteInvite(ctx context.Context, req *v1.DeleteInviteRequest) (*utils_v1.EmptyReply, error) {
	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

	err := s.iu.DeleteInvite(ctx, req.InviteId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) ListInvites(ctx context.Context, req *v1.ListInvitesRequest) (*v1.ListInvitesReply, error) {
	var status *enum.InviteStatus
	if req.Status != v1.Status_UNSPECIFIED {
		statusString := enum.InviteStatus(req.Status.String())
		if !statusString.IsValid() {
			return nil, v1.ErrorInvalidRequest("invalid status")
		}
		status = &statusString
	}

	list, err := s.iu.ListInvites(ctx, data.InvitesListFilter{
		Status: status,
		Search: req.Search,
	}, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}
	return &v1.ListInvitesReply{
		Invites:  replyInvites(list.Invites),
		Paginate: list.Paginate,
	}, nil
}

func replyInvite(invite biz.InviteItem) *v1.Invite {
	return &v1.Invite{
		Id:        invite.ID,
		Email:     invite.Email,
		Status:    v1.Status(v1.Status_value[invite.Status.Value()]),
		CreatedAt: invite.CreatedAt.Format(time.RFC3339),
		UpdatedAt: invite.UpdatedAt.Format(time.RFC3339),
		User:      invite.User,
	}
}

func replyInvites(invites []biz.InviteItem) []*v1.Invite {
	reply := make([]*v1.Invite, len(invites))
	for i, invite := range invites {
		reply[i] = replyInvite(invite)
	}
	return reply
}
