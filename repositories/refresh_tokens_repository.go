package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/refreshtoken"
	"context"
	"time"
)

type RefreshTokensRepository struct {
	client *ent.Client
}

func NewRefreshTokensRepository(client *ent.Client) *RefreshTokensRepository {
	return &RefreshTokensRepository{client: client}
}

func (r *RefreshTokensRepository) CreateRefreshToken(ctx context.Context, jti string, userID int, expiresAt time.Time) error {
	_, err := r.client.RefreshToken.Create().
		SetID(jti).
		SetUserID(userID).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *RefreshTokensRepository) GetActiveRefreshToken(ctx context.Context, jti string) (*ent.RefreshToken, error) {
	return r.client.RefreshToken.Query().
		Where(
			refreshtoken.IDEQ(jti),
			refreshtoken.RevokedAtIsNil(),
		).
		Only(ctx)
}

func (r *RefreshTokensRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	now := time.Now().UTC()
	_, err := r.client.RefreshToken.Update().
		Where(refreshtoken.IDEQ(jti)).
		SetRevokedAt(now).
		Save(ctx)
	return err
}
