package schema

import (
	"time"

	"gitlab.calendaria.team/services/tenants/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Store holds the schema definition for the Store entity.
type Store struct {
	ent.Schema
}

// Fields of the Store.
func (Store) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("tenant_id").Immutable(),
		field.String("name").NotEmpty(),
		field.String("address").Optional().Default(""),
		field.Float("lat").Optional().Nillable(),
		field.Float("lon").Optional().Nillable(),
		field.String("phone").Optional().Default(""),
		field.String("work_hours").Optional().Default(""),
		field.Bool("is_active").Default(true),
		field.Int64("responsible_id").Optional().Nillable(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Store.
func (Store) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("stores").
			Immutable().
			Required().
			Unique().
			Field("tenant_id"),
	}
}

func (Store) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("tenant_id", "is_active"),
	}
}

func (Store) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.SoftDeleteMixin{},
	}
}
