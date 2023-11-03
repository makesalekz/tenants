package data

import (
	"context"

	tenants_v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/ent/member"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type MembersListFilter struct {
	TenantId int64
}

// MembersRepo
type MembersRepo interface {
	CreateMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error)
	DeleteMember(ctx context.Context, tenantId int64, memberId uuid.UUID) error
	GetMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error)
	GetMember(ctx context.Context, tenantId, userId int64) (*ent.Member, error)
	ListMembers(ctx context.Context, filter MembersListFilter, paginate *tenants_v1.PaginateRequest) ([]*ent.Member, error)
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

func (r *membersRepo) DeleteMember(ctx context.Context, tenantId int64, memberId uuid.UUID) error {
	_, err := r.db.Member.Delete().Where(member.TenantID(tenantId), member.IdentityID(memberId)).Exec(ctx)

	return err
}

func (r *membersRepo) GetMembers(ctx context.Context, tenantId int64, usersIds []int64) ([]*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantId), member.UserIDIn(usersIds...)).All(ctx)
}

func (r *membersRepo) GetMember(ctx context.Context, tenantId, userId int64) (*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantId), member.UserID(userId)).Only(ctx)
}

func (r *membersRepo) ListMembers(ctx context.Context, filter MembersListFilter, paginate *tenants_v1.PaginateRequest) ([]*ent.Member, error) {
	query := r.db.Member.Query().Where(member.TenantID(filter.TenantId))

	if paginate.FromId != 0 {
		query.Where(member.IDGT(paginate.FromId))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	return query.Limit(int(paginate.Limit)).Order(ent.Asc(member.FieldID)).All(ctx)
}

func (r *membersRepo) CountListMembers(ctx context.Context, filter MembersListFilter) (int32, error) {
	query := r.db.Member.Query().Where(member.TenantID(filter.TenantId))

	count, err := query.Count(ctx)

	return int32(count), err
}
