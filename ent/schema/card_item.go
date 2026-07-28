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

// CardItemType is the polymorphic type of a card item.
type CardItemType string

const (
	CardItemTypeText                      CardItemType = "TEXT"
	CardItemTypeTextWithHighlightedSymbol CardItemType = "TEXT_WITH_HIGHLIGHTED_SYMBOL"
	CardItemTypeCode                      CardItemType = "CODE"
	CardItemTypeFormula                   CardItemType = "FORMULA"
	CardItemTypeImage                     CardItemType = "IMAGE"
	CardItemTypeWordWithStress            CardItemType = "WORD_WITH_STRESS"
)

// Values returns all valid CardItemType values.
func (CardItemType) Values() []string {
	return []string{
		string(CardItemTypeText),
		string(CardItemTypeTextWithHighlightedSymbol),
		string(CardItemTypeCode),
		string(CardItemTypeFormula),
		string(CardItemTypeImage),
		string(CardItemTypeWordWithStress),
	}
}

func (t CardItemType) String() string {
	return string(t)
}

func (t CardItemType) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(string(t)))
}

func (t *CardItemType) UnmarshalGQL(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("card item type must be a string")
	}
	*t = CardItemType(s)
	return nil
}

// CardItem holds the schema definition for card content blocks (question/answer parts).
type CardItem struct {
	ent.Schema
}

// Fields of the CardItem.
func (CardItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Enum("type").
			GoType(CardItemType("")).
			Immutable(),
		field.String("text").
			Optional().
			Nillable(),
		field.Int("index").
			Optional().
			Nillable(),
		field.Text("code").
			Optional().
			Nillable(),
		field.String("extension").
			Optional().
			Nillable(),
		field.String("formula").
			Optional().
			Nillable(),
		field.String("image_path").
			Optional().
			Nillable(),
		field.String("width").
			Optional().
			Nillable(),
		field.Bool("shared").
			Default(false),
		field.Int("user_id").
			Positive(),
	}
}

// Edges of the CardItem.
func (CardItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("card_items").
			Unique().
			Required().
			Field("user_id"),
	}
}

// Indexes of the CardItem.
func (CardItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("user_id"),
		index.Fields("shared"),
	}
}

// Annotations of the CardItem.
func (CardItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_items"},
	}
}
