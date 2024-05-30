package biz

import "github.com/google/wire"

var DefaultLanguage = "en"

const (
	QueueEmail = "notifications/email"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewTenantsUsecase,
	NewMembersUsecase,
	NewInvitesUsecase,
	NewGroupsUsecase,
)
