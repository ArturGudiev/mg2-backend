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

// ListUserIDs returns user IDs granted access to the memory node.
func (r *MemoryNodeUsersRepository) ListUserIDs(ctx context.Context, memoryNodeID int) ([]int, error) {
	rows, err := r.client.MemoryNodeUser.Query().
		Where(memorynodeuser.MemoryNodeIDEQ(memoryNodeID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids, nil
}

// DeleteByMemoryNodeID removes all grant rows for a memory node.
func (r *MemoryNodeUsersRepository) DeleteByMemoryNodeID(ctx context.Context, memoryNodeID int) error {
	_, err := r.client.MemoryNodeUser.Delete().
		Where(memorynodeuser.MemoryNodeIDEQ(memoryNodeID)).
		Exec(ctx)
	return err
}

// DeleteLink removes the grant row for one user on a memory node (no-op if missing).
func (r *MemoryNodeUsersRepository) DeleteLink(ctx context.Context, memoryNodeID, userID int) error {
	_, err := r.client.MemoryNodeUser.Delete().
		Where(memorynodeuser.MemoryNodeIDEQ(memoryNodeID), memorynodeuser.UserIDEQ(userID)).
		Exec(ctx)
	return err
}
