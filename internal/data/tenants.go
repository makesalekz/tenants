package data

import (
	"context"

	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/ent/member"
	"gitlab.calendaria.team/services/tenants/ent/tenant"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"

	_ "github.com/lib/pq"
)

type TenantDto struct {
	TenantID int64
	OwnerID  int64
	Name     string
	Type     enum.TenantType
}

type TenantsListFilter struct {
	OwnerID int64
	UserID  int64
}

// TenantsRepo.
type TenantsRepo interface {
	CreateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, *ent.Member, error)
	UpdateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, error)
	DeleteTenant(ctx context.Context, tenantID int64) error
	GetTenant(ctx context.Context, tenantID int64) (*ent.Tenant, error)
	ListTenants(ctx context.Context, filter TenantsListFilter, paginate *utils_v1.PaginateRequest) (
		[]*ent.Tenant, error,
	)
	CountListTenants(ctx context.Context, filter TenantsListFilter) (int32, error)
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

	tenant, err := tx.Tenant.Create().
		SetOwnerID(dto.OwnerID).
		SetName(dto.Name).
		SetType(dto.Type).
		Save(ctx)
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

	count, err := query.Count(ctx)

	return int32(count), err
}
