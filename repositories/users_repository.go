package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/ent/user"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type UsersRepository struct {
	client *ent.Client
}

func NewUsersRepository(client *ent.Client) *UsersRepository {
	return &UsersRepository{client: client}
}

func (r *UsersRepository) GetUser(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Get(ctx, id)
}

func (r *UsersRepository) AddUser(ctx context.Context, name, email, password string) (*ent.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return r.client.User.Create().
		SetName(name).
		SetEmail(email).
		SetPasswordHash(string(hash)).
		SetRole(schema.UserRoleUser).
		Save(ctx)
}

func (r *UsersRepository) GetUserByCredentials(ctx context.Context, email, password string) (*ent.User, error) {
	foundUser, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password)); err != nil {
		return nil, &ent.NotFoundError{}
	}
	return foundUser, nil
}

// SetAllRolesAdmin sets every user's role to admin (one-time migration helper).
func (r *UsersRepository) SetAllRolesAdmin(ctx context.Context) (int, error) {
	return r.client.User.Update().SetRole(schema.UserRoleAdmin).Save(ctx)
}
