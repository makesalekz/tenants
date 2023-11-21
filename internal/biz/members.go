package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
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
	Members  []MemberItem
	Paginate *utils_v1.PaginateReply
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

func (uc *MembersUsecase) CreateMembers(ctx context.Context, usersIds []int64) ([]*ent.Member, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("jwt token is missing")
	}

	tenant, err := uc.tenantsRepo.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return nil, err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return nil, v1.ErrorForbidden("only owner can add members")
	}

	return uc.membersRepo.CreateMembers(ctx, tenant.ID, usersIds)
}

func (uc *MembersUsecase) DeleteMember(ctx context.Context, memberId string) error {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return v1.ErrorUnauthorized("jwt token is missing")
	}

	memberUUID, err := uuid.FromBytes([]byte(memberId))
	if err != nil {
		return v1.ErrorInvalidRequest("invalid member id")
	}

	tenant, err := uc.tenantsRepo.GetTenant(ctx, claims.GetTenantId())
	if err != nil {
		return err
	}

	// TODO: check permissions
	if tenant.OwnerID != claims.GetUserId() {
		return v1.ErrorForbidden("only owner can remove members")
	}

	return uc.membersRepo.DeleteMember(ctx, tenant.ID, memberUUID)
}

func (uc *MembersUsecase) GetMember(ctx context.Context, userId int64) (*ent.Member, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("jwt token is missing")
	}

	return uc.membersRepo.GetMember(ctx, claims.GetTenantId(), userId)
}

func (uc *MembersUsecase) ListMembers(ctx context.Context, paginate *utils_v1.PaginateRequest) (*MembersList, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserTenantRequest() {
		return nil, v1.ErrorUnauthorized("jwt token is missing")
	}

	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	filter := data.MembersListFilter{
		TenantId: claims.GetTenantId(),
	}

	members, err := uc.membersRepo.ListMembers(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.membersRepo.CountListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := utils_v1.PaginateReply{
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
