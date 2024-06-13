package data

import (
	"context"

	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/group"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"

	_ "github.com/lib/pq"
)

type CreateGroupDto struct {
	TenantID    int64
	Name        string
	Description string
}

type UpdateGroupDto struct {
	Name        string
	Description string
}

type GroupsListFilter struct {
	TenantID int64
	Search   string
}

// GroupsRepo.
type GroupsRepo interface {
	CreateGroup(ctx context.Context, actorID int64, dto CreateGroupDto) (*ent.Group, error)
	UpdateGroup(ctx context.Context, group *ent.Group, dto UpdateGroupDto) (*ent.Group, error)
	DeleteGroup(ctx context.Context, group *ent.Group) error
	GetGroup(ctx context.Context, tenantID, groupID int64) (*ent.Group, error)
	ListGroups(
		ctx context.Context, filter GroupsListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
	) ([]*ent.Group, error)
	CountListGroups(ctx context.Context, filter GroupsListFilter) (int32, error)
	AddMembersToGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error
	RemoveMembersFromGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error
}

type groupsRepo struct {
	db *ent.Client
}

// NewGroupsRepo .
func NewGroupsRepo(d *Data) GroupsRepo {
	return &groupsRepo{
		db: d.db,
	}
}

func (r *groupsRepo) CreateGroup(ctx context.Context, actorID int64, dto CreateGroupDto) (*ent.Group, error) {
	return r.db.Group.Create().
		SetTenantID(dto.TenantID).
		SetIdentityID(u_uuid.NewFromActorID(actorID)).
		SetName(dto.Name).
		SetDescription(dto.Description).
		Save(ctx)
}

func (r *groupsRepo) UpdateGroup(ctx context.Context, group *ent.Group, dto UpdateGroupDto) (*ent.Group, error) {
	return group.Update().
		SetName(dto.Name).
		SetDescription(dto.Description).
		Save(ctx)
}

func (r *groupsRepo) DeleteGroup(ctx context.Context, group *ent.Group) error {
	return r.db.Group.DeleteOne(group).Exec(ctx)
}

func (r *groupsRepo) GetGroup(ctx context.Context, tenantID, groupID int64) (*ent.Group, error) {
	return r.db.Group.Query().Where(group.TenantID(tenantID), group.ID(groupID)).Only(ctx)
}

func (r *groupsRepo) ListGroups(
	ctx context.Context, filter GroupsListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) ([]*ent.Group, error) {
	query := r.db.Group.Query().
		Where(group.TenantID(filter.TenantID))

	if filter.Search != "" {
		query.Where(group.NameContainsFold(filter.Search))
	}

	if sort != nil {
		var sortField string

		switch sort.GetField() {
		case "name":
			sortField = group.FieldName
		default:
			sortField = group.FieldID
		}

		queryOrder := ent.Asc(sortField)
		if sort.GetDescending() {
			queryOrder = ent.Desc(sortField)
		}

		query.Order(queryOrder)
	} else {
		if paginate.GetFromId() != 0 {
			query.Where(group.IDGT(paginate.GetFromId()))
		}

		query.Order(ent.Asc(group.FieldID))
	}

	if paginate.GetLimit() == 0 {
		paginate.Limit = 100
	}

	if paginate.GetPage() != 0 {
		query.Offset(int((paginate.GetPage() - 1) * paginate.GetLimit()))
	}

	return query.Limit(int(paginate.GetLimit())).All(ctx)
}

func (r *groupsRepo) CountListGroups(ctx context.Context, filter GroupsListFilter) (int32, error) {
	query := r.db.Group.Query().
		Where(group.TenantID(filter.TenantID))

	if filter.Search != "" {
		query.Where(group.NameContainsFold(filter.Search))
	}

	count, err := query.Count(ctx)

	return int32(count), err
}

func (r *groupsRepo) AddMembersToGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error {
	return group.Update().AddMemberIDs(membersIDs...).Exec(ctx)
}

func (r *groupsRepo) RemoveMembersFromGroup(ctx context.Context, group *ent.Group, membersIDs []int64) error {
	return group.Update().RemoveMemberIDs(membersIDs...).Exec(ctx)
}
