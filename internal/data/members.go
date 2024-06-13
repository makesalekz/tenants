package data

import (
	"context"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/group"
	"gitlab.calendaria.team/services/tenants/ent/member"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"
)

type MembersListFilter struct {
	TenantID       int64
	GroupID        int64
	Search         string
	ExcludeGroupID int64
}

// MembersRepo.
type MembersRepo interface {
	CreateMembers(ctx context.Context, tenantID int64, usersIDs []int64) ([]*ent.Member, error)
	DeleteMember(ctx context.Context, tenantID, memberID int64) error
	GetMembers(ctx context.Context, tenantID int64, identityIDs []uuid.UUID) ([]*ent.Member, error)
	GetMember(ctx context.Context, tenantID, memberID int64) (*ent.Member, error)
	GetMemberByUserID(ctx context.Context, tenantID, userID int64) (*ent.Member, error)
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

func (r *membersRepo) CreateMembers(
	ctx context.Context, tenantID int64, usersIDs []int64,
) ([]*ent.Member, error) {
	membersCreate := make([]*ent.MemberCreate, len(usersIDs))
	for i, userID := range usersIDs {
		membersCreate[i] = r.db.Member.
			Create().
			SetTenantID(tenantID).
			SetUserID(userID).
			SetIdentityID(u_uuid.NewFromActorID(userID))
	}

	return r.db.Member.CreateBulk(membersCreate...).Save(ctx)
}

func (r *membersRepo) DeleteMember(ctx context.Context, tenantID, memberID int64) error {
	_, err := r.db.Member.Delete().Where(member.TenantID(tenantID), member.ID(memberID)).Exec(ctx)

	return err
}

func (r *membersRepo) GetMembers(ctx context.Context, tenantID int64, identityIDs []uuid.UUID) ([]*ent.Member, error) {
	return r.db.Member.Query().Where(
		member.TenantID(tenantID),
		member.Or(
			member.IdentityIDIn(identityIDs...),
			member.HasGroupsWith(group.IdentityIDIn(identityIDs...)),
		),
	).WithGroups().All(ctx)
}

func (r *membersRepo) GetMember(ctx context.Context, tenantID, memberID int64) (*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantID), member.ID(memberID)).WithGroups().Only(ctx)
}

func (r *membersRepo) GetMemberByUserID(ctx context.Context, tenantID, userID int64) (*ent.Member, error) {
	return r.db.Member.Query().Where(member.TenantID(tenantID), member.UserID(userID)).WithGroups().Only(ctx)
}

func (r *membersRepo) ListMembers(ctx context.Context, filter MembersListFilter) ([]*ent.Member, error) {
	query := r.db.Member.Query().
		Where(member.TenantID(filter.TenantID))

	if filter.GroupID != 0 {
		query.Where(member.HasGroupsWith(group.ID(filter.GroupID)))
	} else if filter.ExcludeGroupID != 0 {
		query.Where(member.Not(member.HasGroupsWith(group.ID(filter.ExcludeGroupID))))
	}

	return query.All(ctx)
}

func (r *membersRepo) CountListMembers(ctx context.Context, filter MembersListFilter) (int32, error) {
	query := r.db.Member.Query().
		Where(member.TenantID(filter.TenantID))

	if filter.GroupID != 0 {
		query.Where(member.HasGroupsWith(group.ID(filter.GroupID)))
	} else if filter.ExcludeGroupID != 0 {
		query.Where(member.Not(member.HasGroupsWith(group.ID(filter.ExcludeGroupID))))
	}

	count, err := query.Count(ctx)

	return int32(count), err
}
