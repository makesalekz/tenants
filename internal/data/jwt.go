package data

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

type JwtProcessor struct {
	jwtSecret []byte
}

type TenantClaims struct {
	jwtv4.RegisteredClaims

	userId    *int64
	TenantId  int64    `json:"tenant,omitempty"`
	MemberId  string   `json:"member,omitempty"`
	GroupsIds []string `json:"groups,omitempty"`
}

// NewJwtProcessor .
func NewJwtProcessor(c *Config) (*JwtProcessor, error) {
	secret, err := c.ReadGlobalSecretsFor(context.Background(), "jwt")
	if err != nil {
		return nil, fmt.Errorf("jwt secret not found, error: %w", err)
	}

	return &JwtProcessor{
		jwtSecret: []byte(secret["data"].(string)),
	}, nil
}

func (j *JwtProcessor) GetSecret() []byte {
	return j.jwtSecret
}

func (j *JwtProcessor) GetClaimsFromContext(ctx context.Context) (*TenantClaims, bool) {
	token, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, false
	}

	claims, ok := token.(*TenantClaims)
	if !ok {
		return nil, false
	}

	return claims, true
}

func (j *JwtProcessor) GetUserIdFromContext(ctx context.Context) int64 {
	claims, ok := j.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return 0
	}

	return claims.GetUserId()
}

func (tc *TenantClaims) GetUserId() int64 {
	if tc.userId != nil {
		return *tc.userId
	}

	userId, err := strconv.ParseInt(tc.Subject, 10, 64)
	if err != nil {
		userId = 0
	}
	tc.userId = &userId

	return userId
}

func (tc *TenantClaims) GetTenantId() int64 {
	return tc.TenantId
}

func (tc *TenantClaims) GetIdentities() []string {
	return append(tc.GroupsIds, tc.MemberId)
}

func (tc *TenantClaims) IsUserRequest() bool {
	return tc.Issuer == "iam" && tc.GetUserId() != 0
}

func (tc *TenantClaims) IsUserTenantRequest() bool {
	return tc.IsUserRequest() && tc.GetTenantId() != 0
}
