package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"

	"github.com/google/uuid"
)

type InvitesService struct {
	v1.UnimplementedInvitesServer

	tu *biz.TenantsUsecase
	iu *biz.InvitesUsecase
}

func NewInvitesService(
	tu *biz.TenantsUsecase,
	iu *biz.InvitesUsecase,
) *InvitesService {
	return &InvitesService{
		tu: tu,
		iu: iu,
	}
}

func (s *InvitesService) CreateInvites(ctx context.Context, req *v1.CreateInvitesRequest) (*v1.ListInvitesReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	if len(req.Emails) == 0 {
		return nil, v1.ErrorInvalidRequest("emails are empty")
	}

	invites, err := s.iu.CreateInvites(ctx, tenantId, req.Emails)
	if err != nil {
		return nil, err
	}

	return &v1.ListInvitesReply{
		Invites: replyInvites(invites),
	}, nil
}

func (s *InvitesService) CancelInvite(ctx context.Context, req *v1.InviteRequest) (*utils_v1.EmptyReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

	_, err := s.iu.CancelInvite(ctx, tenantId, req.InviteId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) DeleteInvite(ctx context.Context, req *v1.InviteRequest) (*utils_v1.EmptyReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

	err := s.iu.DeleteInvite(ctx, tenantId, req.InviteId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) ListInvites(ctx context.Context, req *v1.ListInvitesRequest) (*v1.ListInvitesReply, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	var status *enum.InviteStatus
	if req.Status != v1.Status_UNSPECIFIED {
		statusString := enum.InviteStatus(req.Status.String())
		if !statusString.IsValid() {
			return nil, v1.ErrorInvalidRequest("invalid status")
		}
		status = &statusString
	}

	list, err := s.iu.ListInvites(ctx, data.InvitesListFilter{
		TenantId: tenantId,
		Status:   status,
		Search:   req.Search,
	}, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}
	return &v1.ListInvitesReply{
		Invites:  replyInvites(list.Invites),
		Paginate: list.Paginate,
	}, nil
}

func (s *InvitesService) AcceptInvite(ctx context.Context, req *v1.InviteCodeRequest) (*v1.TenantReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	if req.Code == "" {
		return nil, v1.ErrorInvalidRequest("code is empty")
	}

	code, err := uuid.Parse(req.Code)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid code")
	}

	invite, err := s.iu.AcceptInvite(ctx, req.InviteId, actorId, code)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("invite not found")
		}
		if ent.IsConstraintError(err) {
			return nil, v1.ErrorInvalidRequest("member already exists")
		}
		return nil, err
	}

	tenant, err := s.tu.GetTenant(ctx, invite.TenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}

	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *InvitesService) ShownInvite(ctx context.Context, req *v1.InviteCodeRequest) (*v1.TenantReply, error) {
	if req.Code == "" {
		return nil, v1.ErrorInvalidRequest("code is empty")
	}

	code, err := uuid.Parse(req.Code)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid code")
	}

	invite, err := s.iu.UpdateInvite(ctx, req.InviteId, code, enum.Shown)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("invite not found")
		}
		return nil, err
	}

	tenant, err := s.tu.GetTenant(ctx, invite.TenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}

	return &v1.TenantReply{
		Tenant: replyTenant(tenant),
	}, nil
}

func (s *InvitesService) DeclineInvite(ctx context.Context, req *v1.InviteCodeRequest) (*utils_v1.EmptyReply, error) {
	if req.Code == "" {
		return nil, v1.ErrorInvalidRequest("code is empty")
	}

	code, err := uuid.Parse(req.Code)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid code")
	}

	_, err = s.iu.UpdateInvite(ctx, req.InviteId, code, enum.Declined)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("invite not found")
		}
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
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
