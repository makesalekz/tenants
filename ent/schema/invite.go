package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"gitlab.calendaria.team/services/tenants/ent/enum"
)

// Invite holds the schema definition for the Invite entity.
type Invite struct {
	ent.Schema
}

// Fields of the Invite.
func (Invite) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id").Immutable(),
		field.UUID("code", uuid.UUID{}).Immutable(),
		field.String("email").Immutable(),
		field.Int64("user_id").Nillable().Optional(),
		field.Enum("status").GoType(enum.InviteStatus("")).Default(enum.New.Value()),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now),
		field.Int64("role_id").Optional(),
		field.String("resource").Optional(),
		field.Int64("resource_id").Optional(),
	}
}

// Edges of the Invite.
func (Invite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("invites").
			Immutable().
			Required().
			Unique().
			Field("tenant_id"),
	}
}

func (Invite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "email").Unique(),
	}
}
