package data

import (
	"context"

	tenants_v1 "tenants/api/tenants/v1"
	"tenants/ent"
	"tenants/ent/member"
	"tenants/ent/tenant"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type TenantDto struct {
	OwnerId int64
	Name    string
}

type TenantsListFilter struct {
	OwnerId  int64
	MemberId *uuid.UUID
}

// TenantsRepo
type TenantsRepo interface {
	CreateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, error)
	UpdateTenant(ctx context.Context, tenantId int64, dto TenantDto) (*ent.Tenant, error)
	DeleteTenant(ctx context.Context, tenantId, ownerId int64) error
	GetTenant(ctx context.Context, tenantId int64) (*ent.Tenant, error)
	ListTenants(ctx context.Context, filter TenantsListFilter, paginate *tenants_v1.PaginateRequest) ([]*ent.Tenant, error)
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

func (r *tenantsRepo) CreateTenant(ctx context.Context, dto TenantDto) (*ent.Tenant, error) {
	return r.db.Tenant.Create().
		SetOwnerID(dto.OwnerId).
		SetName(dto.Name).
		Save(ctx)
}

func (r *tenantsRepo) UpdateTenant(ctx context.Context, tenantId int64, dto TenantDto) (*ent.Tenant, error) {
	return r.db.Tenant.UpdateOneID(tenantId).
		Where(tenant.OwnerID(dto.OwnerId)).
		SetName(dto.Name).
		Save(ctx)
}

func (r *tenantsRepo) DeleteTenant(ctx context.Context, tenantId, ownerId int64) error {
	return r.db.Tenant.DeleteOneID(tenantId).Where(tenant.OwnerID(ownerId)).Exec(ctx)
}

func (r *tenantsRepo) GetTenant(ctx context.Context, tenantId int64) (*ent.Tenant, error) {
	return r.db.Tenant.Get(ctx, tenantId)
}

func (r *tenantsRepo) ListTenants(ctx context.Context, filter TenantsListFilter, paginate *tenants_v1.PaginateRequest) ([]*ent.Tenant, error) {
	query := r.db.Tenant.Query()

	if filter.MemberId != nil {
		query.Where(tenant.HasMembersWith(member.IdentityID(*filter.MemberId)))
	}

	if filter.OwnerId != 0 {
		query.Where(tenant.OwnerID(filter.OwnerId))
	}

	if paginate.FromId != 0 {
		query.Where(tenant.IDGT(paginate.FromId))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	return query.Limit(int(paginate.Limit)).Order(ent.Asc(tenant.FieldID)).All(ctx)
}

func (r *tenantsRepo) CountListTenants(ctx context.Context, filter TenantsListFilter) (int32, error) {
	query := r.db.Tenant.Query()

	if filter.MemberId != nil {
		query.Where(tenant.HasMembersWith(member.IdentityID(*filter.MemberId)))
	}

	if filter.OwnerId != 0 {
		query.Where(tenant.OwnerID(filter.OwnerId))
	}

	count, err := query.Count(ctx)

	return int32(count), err
}
