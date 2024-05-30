package biz

import (
	"github.com/google/wire"
	"gitlab.calendaria.team/services/utils/v1/nats"
)

const (
	QueueEmail = "notifications/email"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewTenantsUsecase,
	NewMembersUsecase,
	NewInvitesUsecase,
	NewGroupsUsecase,
	nats.NewQueueManager,
)
