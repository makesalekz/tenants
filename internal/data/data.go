package data

import (
	"context"
	"os"

	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/conf"
	u_config "gitlab.calendaria.team/services/utils/v1/config"
	u_jwt "gitlab.calendaria.team/services/utils/v1/jwt"
	u_dialer "gitlab.calendaria.team/services/utils/v2/dialer"
	u_tracing "gitlab.calendaria.team/services/utils/v2/tracing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	_ "github.com/lib/pq"
	_ "gitlab.calendaria.team/services/tenants/ent/runtime"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	u_config.NewConfig,
	u_jwt.NewJwtProcessor,
	u_dialer.NewServiceDialerManager,
	u_tracing.NewTracer,
	NewNatsClient,
	NewTenantsRepo,
	NewMembersRepo,
	NewInvitesRepo,
	NewGroupsRepo,
	NewIamRemote,
	NewRbacRemote,
)

// Data .
type Data struct {
	log *log.Helper
	db  *ent.Client
}

// NewData .
func NewData(bc *conf.Bootstrap, c *u_config.Config, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	dbDsn := bc.Db // read from local config
	if dbDsn == "" {
		// read from vault
		secret, err := c.ReadSecretsFor(context.Background(), "db-dsn")
		if err != nil {
			l.Fatalf("db dsn not found: %v", err)
			return nil, nil, err
		}
		dbDsn = secret["data"].(string)
	}

	autoMigrate := os.Getenv("AUTOMIGRATE")
	entLogging := os.Getenv("ENT_LOGGING")
	var options []ent.Option
	if entLogging != "" {
		options = append(options, ent.Debug(), ent.Log(l.Debug))
	}

	client, err := ent.Open("postgres", dbDsn, options...)
	if err != nil {
		l.Fatalf("failed opening connection to postgres: %v", err)
		return nil, nil, err
	}

	if autoMigrate != "" {
		if err := client.Schema.Create(context.Background()); err != nil {
			l.Errorf("failed creating schema resources: %v", err)
			return nil, nil, err
		}
	}

	l.Info("Connected to postgres")

	cleanup := func() {
		if err := client.Close(); err != nil {
			l.Error(err)
		}
	}

	return &Data{
		log: log.NewHelper(logger),
		db:  client,
	}, cleanup, nil
}
