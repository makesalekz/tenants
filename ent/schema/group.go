package schema

import (
	"gitlab.calendaria.team/services/tenants/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("identity_id", uuid.New()).Immutable().Unique(),
		field.Int64("tenant_id").Immutable(),
		field.String("name").NotEmpty(),
		field.String("description"),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("groups").
			Immutable().
			Required().
			Unique().
			Field("tenant_id"),
		edge.To("members", Member.Type),
	}
}

func (Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").Unique(),
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
		mixins.SoftDeleteMixin{},
	}
}
