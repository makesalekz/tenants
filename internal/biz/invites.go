package biz

import (
	"context"
	"fmt"
	"time"

	kconfig "github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/notifications/messages"
	rbac_v1 "gitlab.calendaria.team/services/rbac/api/rbac/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_nats "gitlab.calendaria.team/services/utils/v1/nats"
	"gitlab.calendaria.team/services/utils/v2/auth"
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
	tenantsRepo data.TenantsRepo
	invitesRepo data.InvitesRepo
	iam         data.IIamRemote
	rbac        data.IRbacRemote
	config      kconfig.Config
	qm          u_nats.IQueueManager
}

// NewGreeterUsecase new a Greeter usecase.
func NewInvitesUsecase(
	logger log.Logger,
	tenantsRepo data.TenantsRepo,
	invitesRepo data.InvitesRepo,
	iam data.IIamRemote,
	rbac data.IRbacRemote,
	queueManager u_nats.IQueueManager,
	config kconfig.Config,
) (*InvitesUsecase, error) {
	return &InvitesUsecase{
		log:         log.NewHelper(logger),
		tenantsRepo: tenantsRepo,
		invitesRepo: invitesRepo,
		iam:         iam,
		rbac:        rbac,
		qm:          queueManager,
		config:      config,
	}, nil
}

func (uc *InvitesUsecase) CreateInvites(
	ctx context.Context, tenantID int64, appID string, invite *data.InvitesDTO,
) ([]InviteItem, error) {
	reply, err := uc.iam.GetUsers(ctx, &iam_v1.GetUsersRequest{Emails: invite.Emails})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[string]*iam_v1.UserShort)
	for _, user := range reply.GetUsers() {
		usersMap[user.GetEmail()] = user
	}

	emails := make(map[string]struct{}, len(invite.Emails))
	for i := range invite.Emails {
		emails[invite.Emails[i]] = struct{}{}
	}

	dtos := make([]data.InviteDto, 0, len(emails))
	for email := range emails {
		dto := data.InviteDto{
			Email:      email,
			RoleID:     invite.RoleID,
			Resource:   invite.Resource,
			ResourceID: invite.ResourceID,
		}

		if user, ok := usersMap[email]; ok {
			dto.UserID = &user.Id
		}

		dtos = append(dtos, dto)
	}

	invites, err := uc.invitesRepo.CreateInvites(ctx, tenantID, dtos)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("failed to invite person")
	}

	invitesItems := make([]InviteItem, len(invites))
	for i, invite := range invites {
		invitesItems[i] = InviteItem{Invite: invite}

		if user, ok := usersMap[invite.Email]; ok {
			invitesItems[i].User = user
		}
	}

	if appID == "pms" || appID == "admin" {
		go func() {
			processCtx, processCancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
			defer processCancel()

			uc.processInvitations(processCtx, tenantID, invitesItems, invite.Lang)
		}()
	}

	return invitesItems, nil
}

func (uc *InvitesUsecase) CancelInvite(ctx context.Context, tenantID, inviteID int64) (*ent.Invite, error) {
	invite, err := uc.invitesRepo.GetInvite(ctx, tenantID, inviteID)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	return uc.invitesRepo.UpdateInviteStatus(ctx, invite, enum.Canceled)
}

func (uc *InvitesUsecase) DeleteInvite(ctx context.Context, tenantID, inviteID int64) error {
	return uc.invitesRepo.DeleteInvite(ctx, tenantID, inviteID)
}

func (uc *InvitesUsecase) ListInvites(
	ctx context.Context, filter data.InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) (*InvitesList, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	invites, err := uc.invitesRepo.ListInvites(ctx, filter, sort, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.invitesRepo.CountListInvites(ctx, filter)
	if err != nil {
		return nil, err
	}

	usersIDs := make([]int64, 0, len(invites))
	for _, invite := range invites {
		if invite.UserID != nil {
			usersIDs = append(usersIDs, *invite.UserID)
		}
	}

	reply, err := uc.iam.GetUsers(ctx, &iam_v1.GetUsersRequest{Ids: usersIDs})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[int64]*iam_v1.UserShort)
	for _, user := range reply.GetUsers() {
		usersMap[user.GetId()] = user
	}

	invitesItems := make([]InviteItem, len(invites))
	for i, invite := range invites {
		invitesItems[i] = InviteItem{Invite: invite}

		if invite.UserID != nil {
			invitesItems[i].User = usersMap[*invite.UserID]
		}
	}

	return &InvitesList{
		Invites: invitesItems,
		Paginate: &utils_v1.PaginateReply{
			Total: &total,
		},
	}, nil
}

func (uc *InvitesUsecase) AcceptInvite(ctx context.Context, actorID, inviteID int64, code uuid.UUID) (
	*ent.Invite, error,
) {
	invite, err := uc.invitesRepo.GetInviteByCode(ctx, inviteID, code)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	invite, tenantMember, err := uc.invitesRepo.AcceptInvite(ctx, actorID, invite)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("invite not found")
		}
		if ent.IsConstraintError(err) {
			return nil, v1.ErrorInvalidRequest("member already exists")
		}

		return nil, v1.ErrorDatabaseQuery("failed to accept invite, err %s", err.Error())
	}

	if invite.RoleID != 0 {
		ctx = auth.AppendAuthIds(ctx, tenantMember.UserID, tenantMember.TenantID)

		err = uc.rbac.AssignRoles(
			ctx, &rbac_v1.AssignRoleRequest{
				IdentityId: tenantMember.IdentityID.String(),
				RoleId:     invite.RoleID,
				Resource: &rbac_v1.Resource{
					Type: invite.Resource,
					Id:   invite.ResourceID,
				},
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return invite, nil
}

func (uc *InvitesUsecase) UpdateInvite(
	ctx context.Context, inviteID int64, code uuid.UUID, status enum.InviteStatus,
) (*ent.Invite, error) {
	if status != enum.Shown && status != enum.Declined {
		return nil, v1.ErrorInvalidRequest("invalid status")
	}

	invite, err := uc.invitesRepo.GetInviteByCode(ctx, inviteID, code)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	return uc.invitesRepo.UpdateInviteStatus(ctx, invite, status)
}

func (uc *InvitesUsecase) processInvitations(
	ctx context.Context, tenantID int64, invitesItems []InviteItem, lang string,
) {
	uc.log.Infof("[processInvitations] started with invitesItems: %v", invitesItems)

	tenant, err := uc.tenantsRepo.GetTenant(ctx, tenantID)
	if err != nil {
		uc.log.Errorf("[processInvitations] tenant not found: %v", err)
		return
	}

	owner, err := uc.iam.GetUser(ctx, tenant.OwnerID)
	if err != nil {
		uc.log.Errorf("[processInvitations] owner not found: %v", err)
		return
	}

	baseURL, err := uc.config.Value("INVITE_BASE_URL").String()
	if err != nil {
		uc.log.Errorf("[processInvitations] INVITE_BASE_URL is not provided: %v", err)
		return
	}

	queue := uc.qm.GetRemote(QueueEmail)

	for _, inviteItem := range invitesItems {
		inviteURL := buildInviteLine(baseURL, inviteItem.ID, inviteItem.Invite.Code.String(), lang)
		emailDetailData := map[string]string{
			"InviteLink":    inviteURL,
			"WorkspaceName": tenant.Name,
			"InvitedBy":     owner.GetName(),
		}

		if inviteItem.User != nil && inviteItem.User.GetName() != "" {
			emailDetailData["UserName"] = inviteItem.User.GetName()
		}

		emailDetails := messages.EmailDetails{
			Language: lang,
			Type:     "invite",
			Emails:   []string{inviteItem.Email},
			Data:     emailDetailData,
		}

		queue.Pub(emailDetails)
		uc.log.Infof(
			"[processInvitations] email sent to queue %s [%s] with details %T", QueueEmail, inviteItem.Email,
			emailDetails,
		)
	}
}

func buildInviteLine(baseURL string, inviteID int64, code, lang string) string {
	return fmt.Sprintf("%s/%s/a/%d/%s", baseURL, lang, inviteID, code)
}
