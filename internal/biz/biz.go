package biz

import (
	"github.com/google/wire"

	u_nats "gitlab.calendaria.team/services/utils/v4/nats"
)

const (
	QueueEmail = "notifications.email"

	DefaultTimeout = 30
)

// ProviderSet is biz providers.
//
//nolint:gochecknoglobals // global variable, used in wire
var ProviderSet = wire.NewSet(
	NewTenantsUsecase,
	NewMembersUsecase,
	NewInvitesUsecase,
	NewGroupsUsecase,
	NewStoresUsecase,
	u_nats.NewQueueManager,
)
