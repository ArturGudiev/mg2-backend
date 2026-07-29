package schema

import (
	"fmt"
	"io"
	"strconv"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UsageType describes how a card is practiced.
type UsageType string

const (
	UsageTypeActive       UsageType = "active"
	UsageTypePassive      UsageType = "passive"
	UsageTypeTransitional UsageType = "transitional"
	UsageTypeCommon       UsageType = "common"
)

// Values returns all valid UsageType values.
func (UsageType) Values() []string {
	return []string{
		string(UsageTypeActive),
		string(UsageTypePassive),
		string(UsageTypeTransitional),
		string(UsageTypeCommon),
	}
}

func (t UsageType) String() string {
	return string(t)
}

func (t UsageType) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(string(t)))
}

func (t *UsageType) UnmarshalGQL(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("usage type must be a string")
	}
	*t = UsageType(s)
	return nil
}

// Card holds the schema definition for the Card entity.
// Question and answer reference CardItem IDs from the card_items table.
type Card struct {
	ent.Schema
}

// Fields of the Card.
func (Card) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Ints("question").
			Default([]int{}),
		field.Ints("answer").
			Default([]int{}),
		field.Ints("parent_nodes").
			Default([]int{}),
		field.Int("used").
			Default(0),
		field.Int("needed").
			Default(0),
		field.Int("count").
			Default(0),
		field.Int("reverse_count").
			Default(0),
		field.Enum("usage_type").
			GoType(UsageType("")).
			Default(string(UsageTypeCommon)),
		field.Bool("shared").
			Default(false),
		field.Int("user_id").
			Positive(),
	}
}

// Edges of the Card.
func (Card) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("cards").
			Unique().
			Required().
			Field("user_id"),
		edge.To("user_counts", CardUserCount.Type),
		edge.To("card_users", CardUser.Type),
	}
}

// Indexes of the Card.
func (Card) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("usage_type"),
		index.Fields("user_id"),
		index.Fields("shared"),
	}
}

// Annotations of the Card.
func (Card) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cards"},
	}
}
