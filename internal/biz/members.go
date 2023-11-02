package biz

import (
	"context"

	iam_v1 "iam/api/iam/v1"
	v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type MemberItem struct {
	*ent.Member

	User *iam_v1.UserShort
}

type MembersList struct {
	Members  []MemberItem
	Paginate *v1.PaginateReply
}

// MembersUsecase is a Greeter usecase.
type MembersUsecase struct {
	log         *log.Helper
	jwt         *data.JwtProcessor
	dialer      *data.Dialer
	tenantsRepo data.TenantsRepo
	membersRepo data.MembersRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewMembersUsecase(
	logger log.Logger,
	jwt *data.JwtProcessor,
	dialer *data.Dialer,
	tenantsRepo data.TenantsRepo,
	membersRepo data.MembersRepo,
) (*MembersUsecase, error) {
	return &MembersUsecase{
		log:         log.NewHelper(logger),
		jwt:         jwt,
		dialer:      dialer,
		tenantsRepo: tenantsRepo,
		membersRepo: membersRepo,
	}, nil
}

func (uc *MembersUsecase) CreateMembers(ctx context.Context, teamId int64, usersIds []int64) ([]*ent.Member, error) {
	ownerId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("jwt token is missing")
	}

	team, err := uc.tenantsRepo.GetTenant(ctx, teamId)
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if team.OwnerID != ownerId {
		return nil, v1.ErrorForbidden("only owner can add members")
	}

	return uc.membersRepo.CreateMembers(ctx, teamId, usersIds)
}

func (uc *MembersUsecase) DeleteMember(ctx context.Context, teamId, userId int64) error {
	ownerId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return v1.ErrorUnauthorized("jwt token is missing")
	}

	team, err := uc.tenantsRepo.GetTenant(ctx, teamId)
	if err != nil {
		return err
	}

	// TODO: check permissions
	if team.OwnerID != ownerId {
		return v1.ErrorForbidden("only owner can remove members")
	}

	return uc.membersRepo.DeleteMember(ctx, teamId, userId)
}

func (uc *MembersUsecase) GetMembers(ctx context.Context, teamId int64, usersIds []int64) ([]*ent.Member, error) {
	return uc.membersRepo.GetMembers(ctx, teamId, usersIds)
}

func (uc *MembersUsecase) GetMember(ctx context.Context, teamId, userId int64) (*ent.Member, error) {
	return uc.membersRepo.GetMember(ctx, teamId, userId)
}

func (uc *MembersUsecase) GetOwnMember(ctx context.Context, teamId int64) (*ent.Member, error) {
	userId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("jwt token is missing")
	}

	return uc.membersRepo.GetMember(ctx, teamId, userId)
}

func (uc *MembersUsecase) ListMembers(ctx context.Context, filter data.MembersListFilter, paginate *v1.PaginateRequest) (*MembersList, error) {
	if filter.TenantId == 0 {
		return nil, v1.ErrorInvalidRequest("teamId is required")
	}

	if paginate == nil {
		paginate = &v1.PaginateRequest{}
	}

	members, err := uc.membersRepo.ListMembers(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.membersRepo.CountListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := v1.PaginateReply{
		Total: &total,
	}

	if len(members) == int(paginate.Limit) {
		paginateReply.FromId = &members[len(members)-1].ID
	}

	usersIds := make([]int64, len(members))
	for i, member := range members {
		usersIds[i] = member.UserID
	}

	usersClient, err := uc.dialer.Users(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("dialer.Users: %s", err.Error())
	}

	reply, err := usersClient.GetUsers(ctx, &iam_v1.GetUsersRequest{Ids: usersIds})
	if err != nil {
		return nil, v1.ErrorServiceFailed("usersClient.GetUsers: %s", err.Error())
	}

	users := make(map[int64]*iam_v1.UserShort)
	for _, user := range reply.GetUsers() {
		users[user.Id] = user
	}

	membersItems := make([]MemberItem, len(members))
	for i, member := range members {
		membersItems[i] = MemberItem{
			Member: member,
			User:   users[member.UserID],
		}
	}

	return &MembersList{
		Members:  membersItems,
		Paginate: &paginateReply,
	}, nil
}
