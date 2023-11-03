package schema

import (
	"tenants/ent/mixins"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Member holds the schema definition for the Member entity.
type Member struct {
	ent.Schema
}

// Fields of the Member.
func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("identity_id", uuid.New()).Immutable(),
		field.Int64("tenant_id").Immutable(),
		field.Int64("user_id").Immutable(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

// Edges of the Member.
func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("members").
			Immutable().
			Required().
			Unique().
			Field("tenant_id"),
	}
}

func (Member) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id").Unique(),
	}
}

func (Member) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.SoftDeleteMixin{},
	}
}
