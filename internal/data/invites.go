package data

import (
	"context"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/ent/invite"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_uuid "gitlab.calendaria.team/services/utils/v2/uuid"
)

type InviteDto struct {
	Email  string
	UserID *int64
}

type InvitesListFilter struct {
	TenantID int64
	Search   string
	Status   *enum.InviteStatus
}

// InvitesRepo.
type InvitesRepo interface {
	CreateInvites(ctx context.Context, tenantID int64, dtos []InviteDto) ([]*ent.Invite, error)
	GetInvite(ctx context.Context, tenantID, inviteID int64) (*ent.Invite, error)
	GetInviteByCode(ctx context.Context, inviteID int64, code uuid.UUID) (*ent.Invite, error)
	UpdateInviteStatus(ctx context.Context, invite *ent.Invite, status enum.InviteStatus) (*ent.Invite, error)
	AcceptInvite(ctx context.Context, userID int64, invite *ent.Invite) (*ent.Invite, error)
	DeleteInvite(ctx context.Context, tenantID, inviteID int64) error
	ListInvites(
		ctx context.Context, filter InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
	) ([]*ent.Invite, error)
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

func (r *invitesRepo) CreateInvites(ctx context.Context, tenantID int64, dtos []InviteDto) (
	[]*ent.Invite, error,
) {
	invitesCreate := make([]*ent.InviteCreate, len(dtos))
	for i, dto := range dtos {
		invitesCreate[i] = r.db.Invite.Create().
			SetTenantID(tenantID).
			// we do not care about uniqueness, it's ok. There is probably no security issues
			SetCode(uuid.New()).
			SetEmail(dto.Email)
		if dto.UserID != nil {
			invitesCreate[i].SetUserID(*dto.UserID)
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

	return r.db.Invite.Query().Where(invite.TenantID(tenantID), invite.EmailIn(emails...)).All(ctx)
}

func (r *invitesRepo) GetInvite(ctx context.Context, tenantID, inviteID int64) (*ent.Invite, error) {
	return r.db.Invite.Query().Where(invite.TenantID(tenantID), invite.ID(inviteID)).First(ctx)
}

func (r *invitesRepo) GetInviteByCode(ctx context.Context, inviteID int64, code uuid.UUID) (*ent.Invite, error) {
	return r.db.Invite.Query().Where(invite.ID(inviteID), invite.Code(code)).First(ctx)
}

func (r *invitesRepo) UpdateInviteStatus(
	ctx context.Context, invite *ent.Invite, status enum.InviteStatus,
) (*ent.Invite, error) {
	return invite.Update().SetStatus(status).SetUpdatedAt(time.Now()).Save(ctx)
}

func (r *invitesRepo) AcceptInvite(ctx context.Context, userID int64, invite *ent.Invite) (*ent.Invite, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := tx.Invite.UpdateOneID(invite.ID).
		SetUserID(userID).
		SetStatus(enum.Accepted).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.Member.Create().
		SetTenantID(invite.TenantID).
		SetUserID(userID).
		SetIdentityID(u_uuid.NewFromActorID(userID)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *invitesRepo) DeleteInvite(ctx context.Context, tenantID, inviteID int64) error {
	_, err := r.db.Invite.Delete().Where(invite.TenantID(tenantID), invite.ID(inviteID)).Exec(ctx)

	return err
}

func (r *invitesRepo) ListInvites(
	ctx context.Context, filter InvitesListFilter, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) ([]*ent.Invite, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantID))

	if filter.Status != nil {
		query.Where(invite.StatusEQ(*filter.Status))
	}

	if filter.Search != "" {
		query.Where(invite.EmailContains(filter.Search))
	}

	if sort != nil {
		var sortField string

		switch sort.GetField() {
		case "email":
			sortField = invite.FieldEmail
		case "status":
			sortField = invite.FieldStatus
		default:
			sortField = invite.FieldID
		}

		queryOrder := ent.Asc(sortField)
		if sort.GetDescending() {
			queryOrder = ent.Desc(sortField)
		}

		query.Order(queryOrder)
	} else {
		if paginate.GetFromId() != 0 {
			query.Where(invite.IDGT(paginate.GetFromId()))
		}

		query.Order(ent.Asc(invite.FieldID))
	}

	if paginate.GetLimit() == 0 {
		paginate.Limit = 100
	}

	if paginate.GetPage() != 0 {
		query.Offset(int((paginate.GetPage() - 1) * paginate.GetLimit()))
	}

	return query.Limit(int(paginate.GetLimit())).All(ctx)
}

func (r *invitesRepo) CountListInvites(ctx context.Context, filter InvitesListFilter) (int32, error) {
	query := r.db.Invite.Query().Where(invite.TenantID(filter.TenantID))

	count, err := query.Count(ctx)

	return int32(count), err
}
