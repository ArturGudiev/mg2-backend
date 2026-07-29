package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CardsPriority groups cards under a named priority bucket on a memory node.
type CardsPriority struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
	Cards  []int  `json:"cards"`
}

// CardsGroup groups cards under a named group on a memory node.
type CardsGroup struct {
	Name  string `json:"name"`
	Cards []int  `json:"cards"`
}

// MemoryNode holds the schema definition for the MemoryNode entity.
type MemoryNode struct {
	ent.Schema
}

// Fields of the MemoryNode.
func (MemoryNode) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Ints("children").
			Default([]int{}),
		field.Ints("parents").
			Default([]int{}),
		field.Ints("cards").
			Default([]int{}),
		field.Strings("aliases").
			Default([]string{}),
		field.JSON("priorities", []CardsPriority{}).
			Optional(),
		field.JSON("groups", []CardsGroup{}).
			Optional(),
		field.Bool("shared").
			Default(false),
		field.Int("user_id").
			Positive(),
	}
}

// Edges of the MemoryNode.
func (MemoryNode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("memory_nodes").
			Unique().
			Required().
			Field("user_id"),
		edge.To("memory_node_users", MemoryNodeUser.Type),
	}
}

// Indexes of the MemoryNode.
func (MemoryNode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("shared"),
	}
}

// Annotations of the MemoryNode.
func (MemoryNode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "memory_nodes"},
	}
}
