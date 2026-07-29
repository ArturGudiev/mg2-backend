package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CardUser links a user to a shared card they use.
type CardUser struct {
	ent.Schema
}

// Fields of the CardUser.
func (CardUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("card_id").
			Positive(),
		field.Int("user_id").
			Positive(),
	}
}

// Edges of the CardUser.
func (CardUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("card", Card.Type).
			Ref("card_users").
			Unique().
			Required().
			Field("card_id"),
		edge.From("user", User.Type).
			Ref("card_users").
			Unique().
			Required().
			Field("user_id"),
	}
}

// Indexes of the CardUser.
func (CardUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("card_id", "user_id").
			Unique(),
		index.Fields("user_id"),
		index.Fields("card_id"),
	}
}

// Annotations of the CardUser.
func (CardUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_users"},
	}
}
