package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CardUserCount stores per-user quiz count for shared cards.
type CardUserCount struct {
	ent.Schema
}

// Fields of the CardUserCount.
func (CardUserCount) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("card_id").
			Positive(),
		field.Int("user_id").
			Positive(),
		field.Int("count").
			Default(0),
	}
}

// Edges of the CardUserCount.
func (CardUserCount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("card", Card.Type).
			Ref("user_counts").
			Unique().
			Required().
			Field("card_id"),
		edge.From("user", User.Type).
			Ref("card_user_counts").
			Unique().
			Required().
			Field("user_id"),
	}
}

// Indexes of the CardUserCount.
func (CardUserCount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("card_id", "user_id").
			Unique(),
		index.Fields("user_id"),
		index.Fields("card_id"),
	}
}

// Annotations of the CardUserCount.
func (CardUserCount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_user_counts"},
	}
}
