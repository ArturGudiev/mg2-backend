package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/carduser"
	"arturgudiev/memoryguard/ent/cardusercount"
	"context"
)

type CardUserCountsRepository struct {
	client *ent.Client
}

func NewCardUserCountsRepository(client *ent.Client) *CardUserCountsRepository {
	return &CardUserCountsRepository{client: client}
}

func (r *CardUserCountsRepository) Get(ctx context.Context, cardID, userID int) (*ent.CardUserCount, error) {
	return r.client.CardUserCount.Query().
		Where(cardusercount.CardIDEQ(cardID), cardusercount.UserIDEQ(userID)).
		Only(ctx)
}

// GetCountsByCardIDs returns a map cardID -> count for the given user.
// Missing rows are omitted (caller should treat as 0).
func (r *CardUserCountsRepository) GetCountsByCardIDs(ctx context.Context, cardIDs []int, userID int) (map[int]int, error) {
	out := make(map[int]int, len(cardIDs))
	if len(cardIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.CardUserCount.Query().
		Where(cardusercount.CardIDIn(cardIDs...), cardusercount.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.CardID] = row.Count
	}
	return out, nil
}

func (r *CardUserCountsRepository) SetCount(ctx context.Context, cardID, userID, count int) (*ent.CardUserCount, error) {
	existing, err := r.Get(ctx, cardID, userID)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		return r.client.CardUserCount.Create().
			SetCardID(cardID).
			SetUserID(userID).
			SetCount(count).
			Save(ctx)
	}
	return r.client.CardUserCount.UpdateOneID(existing.ID).
		SetCount(count).
		Save(ctx)
}

func (r *CardUserCountsRepository) Increment(ctx context.Context, cardID, userID int) (*ent.CardUserCount, error) {
	existing, err := r.Get(ctx, cardID, userID)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		return r.client.CardUserCount.Create().
			SetCardID(cardID).
			SetUserID(userID).
			SetCount(1).
			Save(ctx)
	}
	return r.client.CardUserCount.UpdateOneID(existing.ID).
		SetCount(existing.Count + 1).
		Save(ctx)
}

func (r *CardUserCountsRepository) Decrement(ctx context.Context, cardID, userID int) (*ent.CardUserCount, error) {
	existing, err := r.Get(ctx, cardID, userID)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		return r.client.CardUserCount.Create().
			SetCardID(cardID).
			SetUserID(userID).
			SetCount(0).
			Save(ctx)
	}
	next := existing.Count - 1
	if next < 0 {
		next = 0
	}
	return r.client.CardUserCount.UpdateOneID(existing.ID).
		SetCount(next).
		Save(ctx)
}

type CardUsersRepository struct {
	client *ent.Client
}

func NewCardUsersRepository(client *ent.Client) *CardUsersRepository {
	return &CardUsersRepository{client: client}
}

// EnsureLink creates a card_users row if missing.
func (r *CardUsersRepository) EnsureLink(ctx context.Context, cardID, userID int) error {
	exists, err := r.client.CardUser.Query().
		Where(carduser.CardIDEQ(cardID), carduser.UserIDEQ(userID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = r.client.CardUser.Create().
		SetCardID(cardID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		// Concurrent insert: treat unique violation as success if row now exists.
		if again, e2 := r.client.CardUser.Query().
			Where(carduser.CardIDEQ(cardID), carduser.UserIDEQ(userID)).
			Exist(ctx); e2 == nil && again {
			return nil
		}
	}
	return err
}
