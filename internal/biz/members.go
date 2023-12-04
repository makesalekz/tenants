package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type MemberItem struct {
	*ent.Member

	User *iam_v1.UserShort
}

type MembersList struct {
	Members  []MemberItem
	Paginate *utils_v1.PaginateReply
}

// MembersUsecase is a Greeter usecase.
type MembersUsecase struct {
	log         *log.Helper
	jwt         *jwt.JwtProcessor
	dialer      *data.Dialer
	tenantsRepo data.TenantsRepo
	membersRepo data.MembersRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewMembersUsecase(
	logger log.Logger,
	jwt *jwt.JwtProcessor,
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

func (uc *MembersUsecase) CreateMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error) {
	return uc.membersRepo.CreateMembers(ctx, tenantId, usersIds)
}

func (uc *MembersUsecase) DeleteMember(ctx context.Context, tenantId, memberId int64) error {
	return uc.membersRepo.DeleteMember(ctx, tenantId, memberId)
}

func (uc *MembersUsecase) GetMember(ctx context.Context, tenantId, userId int64) (*ent.Member, error) {
	member, err := uc.membersRepo.GetMember(ctx, tenantId, userId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("member not found")
		}
		return nil, err
	}
	return member, nil
}

func (uc *MembersUsecase) ListMembers(ctx context.Context, filter data.MembersListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) (*MembersList, error) {
	members, err := uc.membersRepo.ListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	total, err := uc.membersRepo.CountListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	usersIds := make([]int64, len(members))
	membersMap := make(map[int64]*ent.Member)
	for i, member := range members {
		usersIds[i] = member.UserID
		membersMap[member.UserID] = member
	}

	usersClient, err := uc.dialer.Users(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("iam: %s", err.Error())
	}

	reply, err := usersClient.GetUsers(ctx, &iam_v1.GetUsersRequest{
		Ids:      usersIds,
		Search:   filter.Search,
		Sort:     sort,
		Paginate: paginate,
	})
	if err != nil {
		return nil, v1.ErrorServiceFailed("iam: %s", err.Error())
	}

	membersItems := make([]MemberItem, len(reply.Users))
	for i, user := range reply.Users {
		membersItems[i] = MemberItem{
			Member: membersMap[user.Id],
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
