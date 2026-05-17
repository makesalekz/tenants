package data

import (
	"context"
	"time"

	"gitlab.calendaria.team/services/tenants/ent/group"
	"gitlab.calendaria.team/services/tenants/ent/invite"

	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/ent/member"
	"gitlab.calendaria.team/services/tenants/ent/tenant"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"

	_ "github.com/lib/pq"
)

type TenantDto struct {
	TenantID   int64
	OwnerID    int64
	Name       string
	Type       enum.TenantType
	ReferredBy *int64
}

type TenantsListFilter struct {
	OwnerID    int64
	UserID     int64
	ReferredBy int64
}

// TenantsRepo.
type TenantsRepo interface {
	CreateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, *ent.Member, error)
	UpdateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, error)
	DeleteTenant(ctx context.Context, tenantID int64) error
	DeleteUsersTenants(ctx context.Context, usersIDs []int64) (int, error)
	GetTenant(ctx context.Context, tenantID int64) (*ent.Tenant, error)
	ListTenants(ctx context.Context, filter TenantsListFilter, paginate *utils_v1.PaginateRequest) (
		[]*ent.Tenant, error,
	)
	CountListTenants(ctx context.Context, filter TenantsListFilter) (int32, error)
	TransferOwnership(ctx context.Context, tenantID int64, newOwnerID int64) (*ent.Tenant, error)
}

type tenantsRepo struct {
	db *ent.Client
}

// NewTenantsRepo .
func NewTenantsRepo(d *Data) TenantsRepo {
	return &tenantsRepo{
		db: d.db,
	}
}

func (r *tenantsRepo) CreateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, *ent.Member, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	create := tx.Tenant.Create().
		SetOwnerID(dto.OwnerID).
		SetName(dto.Name).
		SetType(dto.Type)
	if dto.ReferredBy != nil {
		create.SetReferredBy(*dto.ReferredBy)
	}
	tenant, err := create.Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	member, err := tx.Member.Create().
		SetTenantID(tenant.ID).
		SetUserID(dto.OwnerID).
		SetIdentityID(u_uuid.NewFromActorID(dto.OwnerID)).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	return tenant, member, nil
}

func (r *tenantsRepo) UpdateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, error) {
	return r.db.Tenant.UpdateOneID(dto.TenantID).
		SetName(dto.Name).
		Save(ctx)
}

func (r *tenantsRepo) DeleteTenant(ctx context.Context, tenantID int64) error {
	return r.db.Tenant.DeleteOneID(tenantID).Exec(ctx)
}

func (r *tenantsRepo) DeleteUsersTenants(ctx context.Context, usersIDs []int64) (int, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	tenantIDs, err := tx.Tenant.Query().
		Where(
			tenant.OwnerIDIn(usersIDs...),
			tenant.Type(enum.Personal),
		).
		IDs(ctx)

	_, err = tx.Invite.Delete().
		Where(invite.TenantIDIn(tenantIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	_, err = tx.Member.Delete().
		Where(member.TenantIDIn(tenantIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	_, err = tx.Group.Delete().
		Where(group.TenantIDIn(tenantIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	deletedCount, err := r.db.Tenant.Update().
		Where(tenant.IDIn(tenantIDs...)).
		SetName("").
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return deletedCount, nil
}

func (r *tenantsRepo) GetTenant(ctx context.Context, tenantID int64) (*ent.Tenant, error) {
	return r.db.Tenant.Get(ctx, tenantID)
}

func (r *tenantsRepo) ListTenants(
	ctx context.Context, filter TenantsListFilter, paginate *utils_v1.PaginateRequest,
) ([]*ent.Tenant, error) {
	query := r.db.Tenant.Query()

	if filter.UserID != 0 {
		query.Where(tenant.HasMembersWith(member.UserID(filter.UserID)))
	}

	if filter.OwnerID != 0 {
		query.Where(tenant.OwnerID(filter.OwnerID))
	}

	if filter.ReferredBy != 0 {
		query.Where(tenant.ReferredBy(filter.ReferredBy))
	}

	if paginate.GetFromId() != 0 {
		query.Where(tenant.IDGT(paginate.GetFromId()))
	}

	if paginate.GetLimit() == 0 {
		paginate.Limit = 100
	}

	return query.Limit(int(paginate.GetLimit())).Order(ent.Asc(tenant.FieldID)).All(ctx)
}

func (r *tenantsRepo) CountListTenants(ctx context.Context, filter TenantsListFilter) (int32, error) {
	query := r.db.Tenant.Query()

	if filter.UserID != 0 {
		query.Where(tenant.HasMembersWith(member.UserID(filter.UserID)))
	}

	if filter.OwnerID != 0 {
		query.Where(tenant.OwnerID(filter.OwnerID))
	}

	if filter.ReferredBy != 0 {
		query.Where(tenant.ReferredBy(filter.ReferredBy))
	}

	count, err := query.Count(ctx)

	return int32(count), err
}

func (r *tenantsRepo) TransferOwnership(ctx context.Context, tenantID int64, newOwnerID int64) (*ent.Tenant, error) {
	return r.db.Tenant.UpdateOneID(tenantID).
		SetOwnerID(newOwnerID).
		Save(ctx)
}
