package data

import (
	"context"

	"github.com/google/uuid"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/ent/invite"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	_ "github.com/lib/pq"
)

type InviteDto struct {
	Email  string
	UserId *int64
}

type InvitesListFilter struct {
	TenantId int64
}

// InvitesRepo
type InvitesRepo interface {
	CreateInvites(ctx context.Context, tenantId int64, dtos []InviteDto) ([]*ent.Invite, error)
	GetInvite(ctx context.Context, tenantId, inviteId int64) (*ent.Invite, error)
	UpdateInviteStatus(ctx context.Context, invite *ent.Invite, status enum.InviteStatus) (*ent.Invite, error)
	DeleteInvite(ctx context.Context, tenantId, inviteId int64) error
	ListInvites(ctx context.Context, filter InvitesListFilter, paginate *utils_v1.PaginateRequest) ([]*ent.Invite, error)
	CountListInvites(ctx context.Context, filter InvitesListFilter) (int32, error)
}

type invitesRepo struct {
	db *ent.Client
}

// NewInvitesRepo .
func NewInvitesRepo(d *Data) InvitesRepo {
	return &invitesRepo{
		db: d.db,
	}
}

func (r *invitesRepo) CreateInvites(ctx context.Context, tenantId int64, dtos []InviteDto) ([]*ent.Invite, error) {
	invitesCreate := make([]*ent.InviteCreate, len(dtos))
	for i, dto := range dtos {
		invitesCreate[i] = r.db.Invite.Create().SetTenantID(tenantId).SetCode(uuid.New()).SetEmail(dto.Email)
		if dto.UserId != nil {
			invitesCreate[i].SetUserID(*dto.UserId)
		}
	}

	err := r.db.Invite.CreateBulk(invitesCreate...).
		OnConflictColumns(invite.FieldTenantID, invite.FieldEmail).
		UpdateStatus().
		UpdateUpdatedAt().
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(dtos))
	for i, dto := range dtos {
		emails[i] = dto.Email
	}

	return r.db.Invite.Query().Where(invite.TenantID(tenantId), invite.EmailIn(emails...)).All(ctx)
}

func (r *invitesRepo) GetInvite(ctx context.Context, tenantId, inviteId int64) (*ent.Invite, error) {
	return r.db.Invite.Query().Where(invite.TenantID(tenantId), invite.ID(inviteId)).First(ctx)
}

func (r *invitesRepo) UpdateInviteStatus(ctx context.Context, invite *ent.Invite, status enum.InviteStatus) (*ent.Invite, error) {
	return invite.Update().SetStatus(status).Save(ctx)
}

func (r *invitesRepo) DeleteInvite(ctx context.Context, tenantId, inviteId int64) error {
	_, err := r.db.Invite.Delete().Where(invite.TenantID(tenantId), invite.ID(inviteId)).Exec(ctx)

	return err
}

func (r *invitesRepo) ListInvites(ctx context.Context, filter InvitesListFilter, paginate *utils_v1.PaginateRequest) ([]*ent.Invite, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantId))

	if paginate.FromId != 0 {
		query.Where(invite.IDGT(paginate.FromId))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	return query.Limit(int(paginate.Limit)).Order(ent.Asc(invite.FieldID)).All(ctx)
}

func (r *invitesRepo) CountListInvites(ctx context.Context, filter InvitesListFilter) (int32, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantId))

	count, err := query.Count(ctx)

	return int32(count), err
}
