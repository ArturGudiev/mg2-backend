package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/card"
	"arturgudiev/memoryguard/ent/carditem"
	"arturgudiev/memoryguard/ent/carduser"
	"arturgudiev/memoryguard/ent/cardusercount"
	"arturgudiev/memoryguard/ent/memorynode"
	"arturgudiev/memoryguard/ent/memorynodeuser"
	"arturgudiev/memoryguard/ent/refreshtoken"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/ent/user"
	"arturgudiev/memoryguard/ent/verificationcode"
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const verificationCodeTTL = 15 * time.Minute

type UsersRepository struct {
	client *ent.Client
}

func NewUsersRepository(client *ent.Client) *UsersRepository {
	return &UsersRepository{client: client}
}

func (r *UsersRepository) GetUser(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Get(ctx, id)
}

func (r *UsersRepository) DeleteUser(ctx context.Context, id int) error {
	if _, err := r.client.User.Get(ctx, id); err != nil {
		return err
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}

	if err := deleteUserTx(ctx, tx, id); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func deleteUserTx(ctx context.Context, tx *ent.Tx, id int) error {
	cardIDs, err := tx.Card.Query().Where(card.UserIDEQ(id)).IDs(ctx)
	if err != nil {
		return err
	}
	nodeIDs, err := tx.MemoryNode.Query().Where(memorynode.UserIDEQ(id)).IDs(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.CardUser.Delete().Where(carduser.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.CardUserCount.Delete().Where(cardusercount.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.MemoryNodeUser.Delete().Where(memorynodeuser.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if len(cardIDs) > 0 {
		if _, err := tx.CardUser.Delete().Where(carduser.CardIDIn(cardIDs...)).Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.CardUserCount.Delete().Where(cardusercount.CardIDIn(cardIDs...)).Exec(ctx); err != nil {
			return err
		}
	}
	if len(nodeIDs) > 0 {
		if _, err := tx.MemoryNodeUser.Delete().Where(memorynodeuser.MemoryNodeIDIn(nodeIDs...)).Exec(ctx); err != nil {
			return err
		}
	}
	if _, err := tx.Card.Delete().Where(card.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.CardItem.Delete().Where(carditem.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.MemoryNode.Delete().Where(memorynode.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.RefreshToken.Delete().Where(refreshtoken.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.VerificationCode.Delete().Where(verificationcode.UserIDEQ(id)).Exec(ctx); err != nil {
		return err
	}
	return tx.User.DeleteOneID(id).Exec(ctx)
}

func (r *UsersRepository) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.EmailEqualFold(strings.TrimSpace(email))).
		First(ctx)
}

func (r *UsersRepository) GetUserByLoginOrEmail(ctx context.Context, loginOrEmail string) (*ent.User, error) {
	identifier := strings.TrimSpace(loginOrEmail)
	return r.client.User.Query().
		Where(
			user.Or(
				user.LoginEqualFold(identifier),
				user.EmailEqualFold(identifier),
			),
		).
		First(ctx)
}

func (r *UsersRepository) AddUser(ctx context.Context, name, login, email, password string, verified bool) (*ent.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	login, err = r.resolveLogin(ctx, login, email)
	if err != nil {
		return nil, err
	}
	return r.client.User.Create().
		SetName(name).
		SetLogin(login).
		SetEmail(email).
		SetPasswordHash(string(hash)).
		SetRole(schema.UserRoleUser).
		SetVerified(verified).
		Save(ctx)
}

func (r *UsersRepository) UpdateUnverifiedRegistration(ctx context.Context, userID int, name, login, password string) (*ent.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	upd := r.client.User.UpdateOneID(userID).
		SetName(name).
		SetPasswordHash(string(hash))
	if login != "" {
		upd = upd.SetLogin(login)
	}
	return upd.Save(ctx)
}

func (r *UsersRepository) SetVerificationCode(ctx context.Context, userID int, code string) error {
	expires := time.Now().UTC().Add(verificationCodeTTL)
	existing, err := r.client.VerificationCode.Query().
		Where(verificationcode.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
		_, err = r.client.VerificationCode.Create().
			SetUserID(userID).
			SetCode(code).
			SetExpiresAt(expires).
			Save(ctx)
		return err
	}
	return r.client.VerificationCode.UpdateOneID(existing.ID).
		SetCode(code).
		SetExpiresAt(expires).
		Exec(ctx)
}

func (r *UsersRepository) GetVerificationCode(ctx context.Context, userID int) (*ent.VerificationCode, error) {
	return r.client.VerificationCode.Query().
		Where(verificationcode.UserIDEQ(userID)).
		Only(ctx)
}

func (r *UsersRepository) MarkVerified(ctx context.Context, userID int) error {
	if err := r.client.User.UpdateOneID(userID).SetVerified(true).Exec(ctx); err != nil {
		return err
	}
	_, err := r.client.VerificationCode.Delete().
		Where(verificationcode.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// GetUserByCredentials finds a user by login or email and verifies the password.
func (r *UsersRepository) GetUserByCredentials(ctx context.Context, loginOrEmail, password string) (*ent.User, error) {
	foundUser, err := r.client.User.Query().
		Where(
			user.Or(
				user.LoginEqualFold(loginOrEmail),
				user.EmailEqualFold(loginOrEmail),
			),
		).
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

func (r *UsersRepository) resolveLogin(ctx context.Context, login, email string) (string, error) {
	base := strings.TrimSpace(login)
	if base == "" {
		base = loginFromEmail(email)
	}
	return r.ensureUniqueLogin(ctx, base, 0)
}

func (r *UsersRepository) ensureUniqueLogin(ctx context.Context, base string, excludeUserID int) (string, error) {
	candidate := base
	for n := 1; ; n++ {
		q := r.client.User.Query().Where(user.LoginEqualFold(candidate))
		if excludeUserID > 0 {
			q = q.Where(user.IDNEQ(excludeUserID))
		}
		exists, err := q.Exist(ctx)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s%d", base, n)
	}
}

func loginFromEmail(email string) string {
	base := strings.ToLower(strings.TrimSpace(email))
	if at := strings.IndexByte(base, '@'); at > 0 {
		base = base[:at]
	}
	base = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-' {
			return r
		}
		return -1
	}, base)
	if base == "" {
		return "user"
	}
	return base
}
