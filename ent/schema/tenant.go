package schema

import (
	"time"

	"gitlab.calendaria.team/services/tenants/ent/enum"
	"gitlab.calendaria.team/services/tenants/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
	ent.Schema
}

// Fields of the Tenant.
func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Immutable(),
		field.Int64("owner_id"),
		field.String("name"),
		field.Int64("referred_by").Optional().Nillable().Immutable(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now),
		field.String("type").GoType(enum.TenantType("")).Immutable().Default(enum.Business.Value()).Annotations(
			entsql.DefaultExpr("'PERSONAL'"),
		),
	}
}

// Edges of the Tenant.
func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", Member.Type),
		edge.To("groups", Group.Type),
		edge.To("invites", Invite.Type),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "type"),
	}
}

func (Tenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.SoftDeleteMixin{},
	}
}
