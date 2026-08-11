package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MemoryNodeUser links a user to a shared memory node they are allowed to use.
type MemoryNodeUser struct {
	ent.Schema
}

// Fields of the MemoryNodeUser.
func (MemoryNodeUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("memory_node_id").
			Positive(),
		field.Int("user_id").
			Positive(),
	}
}

// Edges of the MemoryNodeUser.
func (MemoryNodeUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("memory_node", MemoryNode.Type).
			Ref("memory_node_users").
			Unique().
			Required().
			Field("memory_node_id"),
		edge.From("user", User.Type).
			Ref("memory_node_users").
			Unique().
			Required().
			Field("user_id"),
	}
}

// Indexes of the MemoryNodeUser.
func (MemoryNodeUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("memory_node_id", "user_id").
			Unique(),
		index.Fields("user_id"),
		index.Fields("memory_node_id"),
	}
}

// Annotations of the MemoryNodeUser.
func (MemoryNodeUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "memory_node_users"},
	}
}
