package biz

import (
	"context"

	log "github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_error "gitlab.calendaria.team/services/utils/v1/error"
)

type GroupsList struct {
	Groups   []*ent.Group
	Paginate *utils_v1.PaginateReply
}

// GroupsUsecase is a Greeter usecase.
type GroupsUsecase struct {
	log         *log.Helper
	tenantsRepo data.TenantsRepo
	groupsRepo  data.GroupsRepo
	membersRepo data.MembersRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewGroupsUsecase(
	logger log.Logger,
	tenantsRepo data.TenantsRepo,
	groupsRepo data.GroupsRepo,
	membersRepo data.MembersRepo,
) (*GroupsUsecase, error) {
	return &GroupsUsecase{
		log:         log.NewHelper(log.With(logger, "module", "usecase/users")),
		tenantsRepo: tenantsRepo,
		groupsRepo:  groupsRepo,
		membersRepo: membersRepo,
	}, nil
}

func (uc *GroupsUsecase) CreateGroup(ctx context.Context, actorID int64, dto data.CreateGroupDto) (*ent.Group, error) {
	group, err := uc.groupsRepo.CreateGroup(ctx, actorID, dto)
	if err != nil {
		if u_error.IsUniqueViolation(err) {
			return nil, v1.ErrorResourceAlreadyExists("group with the same name already exists")
		}
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest("failed validation, err %s", err.Error())
		}
		return nil, err
	}

	return group, nil
}

func (uc *GroupsUsecase) UpdateGroup(ctx context.Context, group *ent.Group, dto data.UpdateGroupDto) (
	*ent.Group, error,
) {
	return uc.groupsRepo.UpdateGroup(ctx, group, dto)
}

func (uc *GroupsUsecase) DeleteGroup(ctx context.Context, group *ent.Group) error {
	return uc.groupsRepo.DeleteGroup(ctx, group)
}

func (uc *GroupsUsecase) GetGroup(ctx context.Context, tenantID, groupID int64) (*ent.Group, error) {
	group, err := uc.groupsRepo.GetGroup(ctx, tenantID, groupID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("group not found")
		}
		return nil, err
	}
	return group, nil
}

func (uc *GroupsUsecase) ListGroups(
	ctx context.Context, filter data.GroupsListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) (*GroupsList, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	groups, err := uc.groupsRepo.ListGroups(ctx, filter, sort, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.groupsRepo.CountListGroups(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &GroupsList{
		Groups: groups,
		Paginate: &utils_v1.PaginateReply{
			Total: &total,
		},
	}, nil
}

func (uc *GroupsUsecase) AddMembersToGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error {
	targetMembersIDs, err := uc.membersRepo.GetTenantMembersIDs(ctx, group.TenantID, membersIDs...)
	if err != nil {
		return err
	}

	if len(membersIDs) != len(targetMembersIDs) {
		uc.log.Warnf("some members were not found when adding to the group: %v", membersIDs)
	}

	return uc.groupsRepo.AddMembersToGroup(ctx, group, targetMembersIDs)
}

func (uc *GroupsUsecase) RemoveMembersFromGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error {
	return uc.groupsRepo.RemoveMembersFromGroup(ctx, group, membersIDs)
}
