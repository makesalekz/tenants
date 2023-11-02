package schema

import (
	"tenants/ent/mixins"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Member holds the schema definition for the Member entity.
type Member struct {
	ent.Schema
}

// Fields of the Member.
func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("user_id"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Member.
func (Member) Edges() []ent.Edge {
	return nil
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
