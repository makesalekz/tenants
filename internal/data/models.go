package data

import (
	"errors"

	"gitlab.calendaria.team/services/tenants/ent/enum"
)

const (
	AllThreeConditions = 0b111
	RoleCondition      = 0b001
	NoConditions       = 0b000
)

type InvitesListFilter struct {
	TenantID int64
	Search   string
	Status   *enum.InviteStatus
}

type InviteDto struct {
	Email      string
	UserID     *int64
	RoleID     int64
	Resource   string
	ResourceID int64
}

type InvitesDTO struct {
	Emails     []string
	Lang       string
	RoleID     int64
	Resource   string
	ResourceID int64
}

func (dto *InvitesDTO) Validate() error {
	if len(dto.Emails) == 0 {
		return errors.New("there is no email in invite")
	}

	var bitmask = btoi(dto.RoleID != 0) |
		btoi(dto.Resource != "")<<1 |
		btoi(dto.ResourceID != 0)<<2

	if bitmask != AllThreeConditions && bitmask != NoConditions && bitmask != RoleCondition {
		return errors.New("invalid invite")
	}

	return nil
}
