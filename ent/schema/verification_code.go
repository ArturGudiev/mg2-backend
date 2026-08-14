package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerificationCode stores a short-lived email confirmation code for a user.
type VerificationCode struct {
	ent.Schema
}

// Fields of the VerificationCode.
func (VerificationCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("user_id").
			Positive().
			Unique(),
		field.String("code").
			NotEmpty().
			Sensitive(),
		field.Time("expires_at"),
		field.Time("created_at").
			Default(time.Now().UTC).
			Immutable(),
	}
}

// Edges of the VerificationCode.
func (VerificationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("verification_codes").
			Unique().
			Required().
			Field("user_id"),
	}
}

// Annotations of the VerificationCode.
func (VerificationCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "verification_codes"},
	}
}
