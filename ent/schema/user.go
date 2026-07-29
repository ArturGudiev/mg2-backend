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
)

// UserRole is the authorization role of a user.
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// Values returns all valid UserRole values.
func (UserRole) Values() []string {
	return []string{
		string(UserRoleAdmin),
		string(UserRoleUser),
	}
}

func (r UserRole) String() string {
	return string(r)
}

func (r UserRole) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(string(r)))
}

func (r *UserRole) UnmarshalGQL(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("user role must be a string")
	}
	*r = UserRole(s)
	return nil
}

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("email").
			NotEmpty().
			Unique(),
		field.String("password").
			NotEmpty().
			Sensitive(),
		field.Enum("role").
			GoType(UserRole("")).
			Default(string(UserRoleUser)),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),
		edge.To("memory_nodes", MemoryNode.Type),
		edge.To("cards", Card.Type),
		edge.To("card_items", CardItem.Type),
		edge.To("card_user_counts", CardUserCount.Type),
		edge.To("card_users", CardUser.Type),
		edge.To("memory_node_users", MemoryNodeUser.Type),
	}
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}
