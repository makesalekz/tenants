package data

import (
	"context"

	"github.com/google/uuid"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/group"
	"gitlab.calendaria.team/services/tenants/ent/member"

	_ "github.com/lib/pq"
)

type MembersListFilter struct {
	TenantId int64
	GroupId  int64
	Search   string
}

// MembersRepo
type MembersRepo interface {
	CreateMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error)
	DeleteMember(ctx context.Context, tenantId, memberId int64) error
	GetMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error)
	GetMember(ctx context.Context, tenantId, userId int64) (*ent.Member, error)
	ListMembers(ctx context.Context, filter MembersListFilter) ([]*ent.Member, error)
	CountListMembers(ctx context.Context, filter MembersListFilter) (int32, error)
}

type membersRepo struct {
	db *ent.Client
}

// NewMembersRepo .
func NewMembersRepo(d *Data) MembersRepo {
	return &membersRepo{
		db: d.db,
	}
}

func (r *membersRepo) CreateMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error) {
	membersCreate := make([]*ent.MemberCreate, len(usersIds))
	for i, userId := range usersIds {
		membersCreate[i] = r.db.Member.Create().SetTenantID(tenantId).SetUserID(userId).SetIdentityID(uuid.New())
	}

	return r.db.Member.CreateBulk(membersCreate...).Save(ctx)
}

func (r *membersRepo) DeleteMember(ctx context.Context, tenantId, memberId int64) error {
	_, err := r.db.Member.Delete().Where(member.TenantID(tenantId), member.ID(memberId)).Exec(ctx)

	return err
}

func (r *membersRepo) GetMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantId), member.UserIDIn(usersIds...)).All(ctx)
}

func (r *membersRepo) GetMember(ctx context.Context, tenantId, userId int64) (*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantId), member.UserID(userId)).Only(ctx)
}

func (r *membersRepo) ListMembers(ctx context.Context, filter MembersListFilter) ([]*ent.Member, error) {
	query := r.db.Member.Query().
		Where(member.TenantID(filter.TenantId))

	if filter.GroupId != 0 {
		query.Where(member.HasGroupsWith(group.ID(filter.GroupId)))
	}

	return query.All(ctx)
}

func (r *membersRepo) CountListMembers(ctx context.Context, filter MembersListFilter) (int32, error) {
	query := r.db.Member.Query().
		Where(member.TenantID(filter.TenantId))

	count, err := query.Count(ctx)

	return int32(count), err
}
