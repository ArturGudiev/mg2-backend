package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/memorynodeuser"
	"context"
)

type MemoryNodeUsersRepository struct {
	client *ent.Client
}

func NewMemoryNodeUsersRepository(client *ent.Client) *MemoryNodeUsersRepository {
	return &MemoryNodeUsersRepository{client: client}
}

// EnsureLink creates a memory_node_users row if missing.
func (r *MemoryNodeUsersRepository) EnsureLink(ctx context.Context, memoryNodeID, userID int) error {
	exists, err := r.client.MemoryNodeUser.Query().
		Where(memorynodeuser.MemoryNodeIDEQ(memoryNodeID), memorynodeuser.UserIDEQ(userID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = r.client.MemoryNodeUser.Create().
		SetMemoryNodeID(memoryNodeID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		if again, e2 := r.client.MemoryNodeUser.Query().
			Where(memorynodeuser.MemoryNodeIDEQ(memoryNodeID), memorynodeuser.UserIDEQ(userID)).
			Exist(ctx); e2 == nil && again {
			return nil
		}
	}
	return err
}
