package biz

import (
	"context"

	"github.com/google/uuid"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type MemberItem struct {
	*ent.Member

	User *iam_v1.UserShort
}

type MembersList struct {
	Members  []*MemberItem
	Paginate *utils_v1.PaginateReply
}

// MembersUsecase is a Greeter usecase.
type MembersUsecase struct {
	tenantsRepo data.TenantsRepo
	membersRepo data.MembersRepo
	iam         *data.IamRemote
}

// NewGreeterUsecase new a Greeter usecase.
func NewMembersUsecase(
	tenantsRepo data.TenantsRepo,
	membersRepo data.MembersRepo,
	iam *data.IamRemote,
) (*MembersUsecase, error) {
	return &MembersUsecase{
		tenantsRepo: tenantsRepo,
		membersRepo: membersRepo,
		iam:         iam,
	}, nil
}

func (uc *MembersUsecase) CreateMembers(ctx context.Context, tenantID int64, usersIDs []int64) (
	[]*ent.Member, error,
) {
	return uc.membersRepo.CreateMembers(ctx, tenantID, usersIDs)
}

func (uc *MembersUsecase) DeleteMember(ctx context.Context, tenantID, memberID int64) error {
	return uc.membersRepo.DeleteMember(ctx, tenantID, memberID)
}

func (uc *MembersUsecase) GetMember(ctx context.Context, tenantID, memberID int64) (*MemberItem, error) {
	member, err := uc.membersRepo.GetMember(ctx, tenantID, memberID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("member not found")
		}
		return nil, err
	}

	user, err := uc.iam.GetUser(ctx, member.UserID)
	if err != nil {
		return nil, err
	}

	return &MemberItem{
		Member: member,
		User:   user,
	}, nil
}

func (uc *MembersUsecase) GetShortMembers(ctx context.Context, tenantID int64, identities []string, withGroups bool) (
	[]*ent.Member, error,
) {
	identityUuids := make([]uuid.UUID, len(identities))
	var err error
	for i, identity := range identities {
		identityUuids[i], err = uuid.Parse(identity)
		if err != nil {
			return nil, v1.ErrorInvalidRequest("invalid identity, %s", err.Error())
		}
	}

	members, err := uc.membersRepo.GetMembers(ctx, tenantID, identityUuids, withGroups)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (uc *MembersUsecase) GetMemberByUserID(ctx context.Context, tenantID, userID int64) (*ent.Member, error) {
	member, err := uc.membersRepo.GetMemberByUserID(ctx, tenantID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("member not found")
		}
		return nil, err
	}
	return member, nil
}

func (uc *MembersUsecase) ListMembers(
	ctx context.Context, filter data.MembersListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) (*MembersList, error) {
	members, err := uc.membersRepo.ListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	total, err := uc.membersRepo.CountListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	usersIDs := make([]int64, len(members))
	membersMap := make(map[int64]*ent.Member)
	for i, member := range members {
		usersIDs[i] = member.UserID
		membersMap[member.UserID] = member
	}

	reply, err := uc.iam.ListUsers(
		ctx, &iam_v1.ListUsersRequest{
			Ids:      usersIDs,
			Search:   filter.Search,
			Sort:     sort,
			Paginate: paginate,
		},
	)
	if err != nil {
		return nil, err
	}

	membersItems := make([]*MemberItem, len(reply.GetUsers()))
	for i, user := range reply.GetUsers() {
		membersItems[i] = &MemberItem{
			Member: membersMap[user.GetId()],
			User:   user,
		}
	}

	return &MembersList{
		Members: membersItems,
		Paginate: &utils_v1.PaginateReply{
			Total: &total,
		},
	}, nil
}
