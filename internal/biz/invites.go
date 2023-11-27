package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type InviteItem struct {
	*ent.Invite

	User *iam_v1.UserShort
}

type InvitesList struct {
	Invites  []InviteItem
	Paginate *utils_v1.PaginateReply
}

// InvitesUsecase is a Greeter usecase.
type InvitesUsecase struct {
	log         *log.Helper
	jwt         *jwt.JwtProcessor
	tenantsRepo data.TenantsRepo
	invitesRepo data.InvitesRepo
	iam         *data.IamRemote
}

// NewGreeterUsecase new a Greeter usecase.
func NewInvitesUsecase(
	logger log.Logger,
	jwt *jwt.JwtProcessor,
	tenantsRepo data.TenantsRepo,
	invitesRepo data.InvitesRepo,
	iam *data.IamRemote,
) (*InvitesUsecase, error) {
	return &InvitesUsecase{
		log:         log.NewHelper(logger),
		jwt:         jwt,
		tenantsRepo: tenantsRepo,
		invitesRepo: invitesRepo,
		iam:         iam,
	}, nil
}

func (uc *InvitesUsecase) CreateInvites(ctx context.Context, emails []string) ([]InviteItem, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := uc.tenantsRepo.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can create invites")
	}

	users, err := uc.iam.GetUsers(ctx, nil, emails)
	if err != nil {
		return nil, err
	}

	usersMap := make(map[string]*iam_v1.UserShort)
	for _, user := range users {
		usersMap[user.Email] = user
	}

	dtos := make([]data.InviteDto, len(emails))
	for i, email := range emails {
		dtos[i] = data.InviteDto{
			Email: email,
		}
		if user, ok := usersMap[email]; ok {
			dtos[i].UserId = &user.Id
		}
	}

	invites, err := uc.invitesRepo.CreateInvites(ctx, tenant.ID, dtos)
	if err != nil {
		return nil, err
	}

	invitesItems := make([]InviteItem, len(invites))
	for i, invite := range invites {
		invitesItems[i] = InviteItem{Invite: invite}

		if user, ok := usersMap[invite.Email]; ok {
			invitesItems[i].User = user
		}
	}

	return invitesItems, nil
}

func (uc *InvitesUsecase) UpdateInvite(ctx context.Context, inviteId int64, status enum.InviteStatus) (*ent.Invite, error) {

	uc.log.Debugf("UpdateInvite: %d, %s", inviteId, status)

	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := uc.tenantsRepo.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can remove invites")
	}

	invite, err := uc.invitesRepo.GetInvite(ctx, claims.GetTenantId(), inviteId)
	if err != nil {
		return nil, err
	}

	return uc.invitesRepo.UpdateInviteStatus(ctx, invite, status)
}

func (uc *InvitesUsecase) DeleteInvite(ctx context.Context, inviteId int64) error {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return v1.ErrorUnauthorized("invalid token")
	}

	tenant, err := uc.tenantsRepo.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return v1.ErrorForbidden("only owner can remove invites")
	}

	return uc.invitesRepo.DeleteInvite(ctx, tenant.ID, inviteId)
}

func (uc *InvitesUsecase) ListInvites(ctx context.Context, filter data.InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) (*InvitesList, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	filter.TenantId = claims.GetTenantId()

	invites, err := uc.invitesRepo.ListInvites(ctx, filter, sort, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.invitesRepo.CountListInvites(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := utils_v1.PaginateReply{
		Total: &total,
	}

	if len(invites) == int(paginate.Limit) {
		paginateReply.FromId = &invites[len(invites)-1].ID
	}

	usersIds := make([]int64, 0, len(invites))
	for _, invite := range invites {
		if invite.UserID != nil {
			usersIds = append(usersIds, *invite.UserID)
		}
	}

	users, err := uc.iam.GetUsers(ctx, usersIds, nil)
	if err != nil {
		return nil, err
	}

	usersMap := make(map[int64]*iam_v1.UserShort)
	for _, user := range users {
		usersMap[user.Id] = user
	}

	invitesItems := make([]InviteItem, len(invites))
	for i, invite := range invites {
		invitesItems[i] = InviteItem{Invite: invite}

		if invite.UserID != nil {
			invitesItems[i].User = usersMap[*invite.UserID]
		}
	}

	return &InvitesList{
		Invites:  invitesItems,
		Paginate: &paginateReply,
	}, nil
}
