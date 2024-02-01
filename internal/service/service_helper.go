package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type ServiceHelper struct {
	jwt *jwt.JwtProcessor
}

func NewServiceHelper(
	jwt *jwt.JwtProcessor,
) *ServiceHelper {
	return &ServiceHelper{
		jwt: jwt,
	}
}

func (s *ServiceHelper) GetActorId(ctx context.Context, reqActorId int64) (int64, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId != 0 {
		return actorId, nil
	}

	// TODO: remove getting from context
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if ok && claims.IsUserTenantRequest() {
		return claims.GetUserId(), nil
	}

	if reqActorId != 0 {
		return reqActorId, nil
	}
	return 0, v1.ErrorUnauthorized("invalid token")
}

func (s *ServiceHelper) GetTenantId(ctx context.Context, reqTenantId int64) (int64, error) {
	tenantId := auth.GetTenantIdFromContext(ctx)
	if tenantId != 0 {
		return tenantId, nil
	}

	// TODO: remove getting from context
	claims, ok := s.jwt.GetClaimsFromContext(ctx)
	if ok && claims.IsUserTenantRequest() {
		return claims.GetTenantId(), nil
	}

	if reqTenantId != 0 {
		return reqTenantId, nil
	}
	return 0, v1.ErrorUnauthorized("invalid token")
}
