package biz

import (
	"context"
	"fmt"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/nats"

	"github.com/google/uuid"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/notifications/messages"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
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
	tenantsRepo data.TenantsRepo
	invitesRepo data.InvitesRepo
	iam         *data.IamRemote
	config      *config.Config
	queue       *nats.QueueManager
}

// NewGreeterUsecase new a Greeter usecase.
func NewInvitesUsecase(
	tenantsRepo data.TenantsRepo,
	invitesRepo data.InvitesRepo,
	iam *data.IamRemote,
	queueManager *nats.QueueManager,
) (*InvitesUsecase, error) {
	return &InvitesUsecase{
		tenantsRepo: tenantsRepo,
		invitesRepo: invitesRepo,
		iam:         iam,
		queue:       queueManager,
	}, nil
}

func buildInviteLine(baseUrl string, userId *int64, code string) string {
	//baseUrl + "/a/" + userId + "/" + code
	return fmt.Sprintf("%s/a/%d/%s", baseUrl, *userId, code)
}

func (uc *InvitesUsecase) CreateInvites(ctx context.Context, tenantId int64, emails []string, appId string) ([]InviteItem, error) {
	reply, err := uc.iam.GetUsers(ctx, &iam_v1.GetUsersRequest{Emails: emails})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[string]*iam_v1.UserShort)
	for _, user := range reply.Users {
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

	invites, err := uc.invitesRepo.CreateInvites(ctx, tenantId, dtos)
	if err != nil {
		return nil, err
	}

	invitesItems := make([]InviteItem, len(invites))
	for i, invite := range invites {
		invitesItems[i] = InviteItem{Invite: invite}

		if user, ok := usersMap[invite.Email]; ok {
			invitesItems[i].User = user
		}

		if appId == "pms" || appId == "admin" {
			uc.sendInviteEmail(invite)
		}
	}

	return invitesItems, nil
}

func (uc *InvitesUsecase) sendInviteEmail(invite *ent.Invite) {
	baseUrl, err := uc.config.Value("INVITE_BASE_URL").String()
	if err == nil {
		return
	}
	inviteUrl := buildInviteLine(baseUrl, invite.UserID, invite.Code.String())
	uc.queue.GetRemote(QueueEmail).Pub(messages.EmailDetails{
		Language: "en",
		Type:     "invite",
		Email:    invite.Email,
		Data: map[string]interface{}{
			"InviteLink": inviteUrl,
		},
	})
}

func (uc *InvitesUsecase) CancelInvite(ctx context.Context, tenantId, inviteId int64) (*ent.Invite, error) {
	invite, err := uc.invitesRepo.GetInvite(ctx, tenantId, inviteId)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	return uc.invitesRepo.UpdateInviteStatus(ctx, invite, enum.Canceled)
}

func (uc *InvitesUsecase) DeleteInvite(ctx context.Context, tenantId, inviteId int64) error {
	return uc.invitesRepo.DeleteInvite(ctx, tenantId, inviteId)
}

func (uc *InvitesUsecase) ListInvites(ctx context.Context, filter data.InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) (*InvitesList, error) {
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

	usersIds := make([]int64, 0, len(invites))
	for _, invite := range invites {
		if invite.UserID != nil {
			usersIds = append(usersIds, *invite.UserID)
		}
	}

	reply, err := uc.iam.GetUsers(ctx, &iam_v1.GetUsersRequest{Ids: usersIds})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[int64]*iam_v1.UserShort)
	for _, user := range reply.Users {
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
		Invites: invitesItems,
		Paginate: &utils_v1.PaginateReply{
			Total: &total,
		},
	}, nil
}

func (uc *InvitesUsecase) AcceptInvite(ctx context.Context, inviteId, userId int64, code uuid.UUID) (*ent.Invite, error) {
	invite, err := uc.invitesRepo.GetInviteByCode(ctx, inviteId, code)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	return uc.invitesRepo.AcceptInvite(ctx, userId, invite)
}

func (uc *InvitesUsecase) UpdateInvite(ctx context.Context, inviteId int64, code uuid.UUID, status enum.InviteStatus) (*ent.Invite, error) {
	if status != enum.Shown && status != enum.Declined {
		return nil, v1.ErrorInvalidRequest("invalid status")
	}

	invite, err := uc.invitesRepo.GetInviteByCode(ctx, inviteId, code)
	if err != nil {
		return nil, err
	}

	if invite.Status == enum.Accepted || invite.Status == enum.Declined || invite.Status == enum.Canceled {
		return nil, v1.ErrorForbidden("invite is already accepted or declined")
	}

	return uc.invitesRepo.UpdateInviteStatus(ctx, invite, status)
}
