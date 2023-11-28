package data

import (
	"context"
	"time"

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
	Search   string
	Status   *enum.InviteStatus
}

// InvitesRepo
type InvitesRepo interface {
	CreateInvites(ctx context.Context, tenantId int64, dtos []InviteDto) ([]*ent.Invite, error)
	GetInvite(ctx context.Context, tenantId, inviteId int64) (*ent.Invite, error)
	GetInviteByCode(ctx context.Context, inviteId int64, code uuid.UUID) (*ent.Invite, error)
	UpdateInviteStatus(ctx context.Context, invite *ent.Invite, status enum.InviteStatus) (*ent.Invite, error)
	AcceptInvite(ctx context.Context, userId int64, invite *ent.Invite) (*ent.Invite, error)
	DeleteInvite(ctx context.Context, tenantId, inviteId int64) error
	ListInvites(ctx context.Context, filter InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.Invite, error)
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

func (r *invitesRepo) GetInviteByCode(ctx context.Context, inviteId int64, code uuid.UUID) (*ent.Invite, error) {
	return r.db.Invite.Query().Where(invite.ID(inviteId), invite.Code(code)).First(ctx)
}

func (r *invitesRepo) UpdateInviteStatus(ctx context.Context, invite *ent.Invite, status enum.InviteStatus) (*ent.Invite, error) {
	return invite.Update().SetStatus(status).SetUpdatedAt(time.Now()).Save(ctx)
}

func (r *invitesRepo) AcceptInvite(ctx context.Context, userId int64, invite *ent.Invite) (*ent.Invite, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	updated, err := tx.Invite.UpdateOneID(invite.ID).SetUserID(userId).SetStatus(enum.Accepted).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.Member.Create().SetTenantID(invite.TenantID).SetUserID(userId).SetIdentityID(uuid.New()).Save(ctx)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return updated, nil
}

func (r *invitesRepo) DeleteInvite(ctx context.Context, tenantId, inviteId int64) error {
	_, err := r.db.Invite.Delete().Where(invite.TenantID(tenantId), invite.ID(inviteId)).Exec(ctx)

	return err
}

func (r *invitesRepo) ListInvites(ctx context.Context, filter InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.Invite, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantId))

	if filter.Status != nil {
		query.Where(invite.StatusEQ(*filter.Status))
	}

	if filter.Search != "" {
		query.Where(invite.EmailContains(filter.Search))
	}

	if sort != nil {
		switch sort.Field {
		case "email":
			if sort.Descending {
				query.Order(ent.Desc(invite.FieldEmail))
			} else {
				query.Order(ent.Asc(invite.FieldEmail))
			}
		case "status":
			if sort.Descending {
				query.Order(ent.Desc(invite.FieldStatus))
			} else {
				query.Order(ent.Asc(invite.FieldStatus))
			}
		default: // case "id"
			if sort.Descending {
				query.Order(ent.Desc(invite.FieldID))
			} else {
				query.Order(ent.Asc(invite.FieldID))
			}
		}
	} else {
		if paginate.FromId != 0 {
			query.Where(invite.IDGT(paginate.FromId))
		}

		query.Order(ent.Asc(invite.FieldID))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	if paginate.Page != 0 {
		query.Offset(int((paginate.Page - 1) * paginate.Limit))
	}

	return query.Limit(int(paginate.Limit)).All(ctx)
}

func (r *invitesRepo) CountListInvites(ctx context.Context, filter InvitesListFilter) (int32, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantId))

	count, err := query.Count(ctx)

	return int32(count), err
}
