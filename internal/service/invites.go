package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type InvitesService struct {
	v1.UnimplementedInvitesServer

	jwt *jwt.JwtProcessor
	tu  *biz.TenantsUsecase
	iu  *biz.InvitesUsecase
}

func NewInvitesService(jwt *jwt.JwtProcessor, tu *biz.TenantsUsecase, iu *biz.InvitesUsecase) *InvitesService {
	return &InvitesService{
		jwt: jwt,
		tu:  tu,
		iu:  iu,
	}
}

func (s *InvitesService) CreateInvites(ctx context.Context, req *v1.CreateInvitesRequest) (*v1.ListInvitesReply, error) {
	if len(req.Emails) == 0 {
		return nil, v1.ErrorInvalidRequest("emails is empty")
	}

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
		return nil, v1.ErrorForbidden("only owner can create invites")
	}

	invites, err := s.iu.CreateInvites(ctx, claims.GetTenantId(), req.Emails)
	if err != nil {
		return nil, err
	}

	return &v1.ListInvitesReply{
		Invites: replyInvites(invites),
	}, nil
}

func (s *InvitesService) CancelInvite(ctx context.Context, req *v1.InviteRequest) (*utils_v1.EmptyReply, error) {
	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

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
		return nil, v1.ErrorForbidden("only owner can update invites")
	}

	_, err = s.iu.CancelInvite(ctx, claims.GetTenantId(), req.InviteId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) DeleteInvite(ctx context.Context, req *v1.InviteRequest) (*utils_v1.EmptyReply, error) {
	if req.InviteId == 0 {
		return nil, v1.ErrorInvalidRequest("invite_id is empty")
	}

	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := s.tu.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("tenant not found")
		}
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can remove invites")
	}

	err = s.iu.DeleteInvite(ctx, tenant.ID, req.InviteId)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) ListInvites(ctx context.Context, req *v1.ListInvitesRequest) (*v1.ListInvitesReply, error) {
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
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
		TenantId: claims.GetTenantId(),
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

func (s *InvitesService) AcceptInvite(ctx context.Context, req *v1.InviteCodeRequest) (*utils_v1.EmptyReply, error) {
	if req.Code == "" {
		return nil, v1.ErrorInvalidRequest("code is empty")
	}

	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	code, err := uuid.Parse(req.Code)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid code")
	}

	_, err = s.iu.AcceptInvite(ctx, req.InviteId, claims.GetUserId(), code)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("invite not found")
		}
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *InvitesService) ShownInvite(ctx context.Context, req *v1.InviteCodeRequest) (*v1.InviteShownReply, error) {
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

	return &v1.InviteShownReply{
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
