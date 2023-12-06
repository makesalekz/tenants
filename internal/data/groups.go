package data

import (
	"context"

	"github.com/google/uuid"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/group"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	_ "github.com/lib/pq"
)

type CreateGroupDto struct {
	TenantId    int64
	Name        string
	Description string
}

type UpdateGroupDto struct {
	Name        string
	Description string
}

type GroupsListFilter struct {
	TenantId int64
	Search   string
}

// GroupsRepo
type GroupsRepo interface {
	CreateGroup(ctx context.Context, dto CreateGroupDto) (*ent.Group, error)
	UpdateGroup(ctx context.Context, group *ent.Group, dto UpdateGroupDto) (*ent.Group, error)
	DeleteGroup(ctx context.Context, group *ent.Group) error
	GetGroup(ctx context.Context, tenantId, groupId int64) (*ent.Group, error)
	ListGroups(ctx context.Context, filter GroupsListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.Group, error)
	CountListGroups(ctx context.Context, filter GroupsListFilter) (int32, error)
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

func (r *groupsRepo) CreateGroup(ctx context.Context, dto CreateGroupDto) (*ent.Group, error) {
	return r.db.Group.Create().
		SetTenantID(dto.TenantId).
		SetIdentityID(uuid.New()).
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

func (r *groupsRepo) GetGroup(ctx context.Context, tenantId, groupId int64) (*ent.Group, error) {
	return r.db.Group.Query().Where(group.TenantID(tenantId), group.ID(groupId)).Only(ctx)
}

func (r *groupsRepo) ListGroups(ctx context.Context, filter GroupsListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.Group, error) {
	query := r.db.Group.Query().
		Where(group.TenantID(filter.TenantId))

	if filter.Search != "" {
		query.Where(group.NameContainsFold(filter.Search))
	}

	if sort != nil {
		switch sort.Field {
		case "name":
			if sort.Descending {
				query.Order(ent.Desc(group.FieldName))
			} else {
				query.Order(ent.Asc(group.FieldName))
			}
		default: // case "id"
			if sort.Descending {
				query.Order(ent.Desc(group.FieldID))
			} else {
				query.Order(ent.Asc(group.FieldID))
			}
		}
	} else {
		if paginate.FromId != 0 {
			query.Where(group.IDGT(paginate.FromId))
		}

		query.Order(ent.Asc(group.FieldID))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	if paginate.Page != 0 {
		query.Offset(int((paginate.Page - 1) * paginate.Limit))
	}

	return query.Limit(int(paginate.Limit)).All(ctx)
}

func (r *groupsRepo) CountListGroups(ctx context.Context, filter GroupsListFilter) (int32, error) {
	query := r.db.Group.Query().
		Where(group.TenantID(filter.TenantId))

	if filter.Search != "" {
		query.Where(group.NameContainsFold(filter.Search))
	}

	count, err := query.Count(ctx)

	return int32(count), err
}
