package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/memorynode"
	"arturgudiev/memoryguard/ent/predicate"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"context"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

type MemoryNodesRepository struct {
	client *ent.Client
}

func NewMemoryNodesRepository(client *ent.Client) *MemoryNodesRepository {
	return &MemoryNodesRepository{client: client}
}

func accessibleMemoryNode(userID int) predicate.MemoryNode {
	return memorynode.Or(memorynode.UserIDEQ(userID), memorynode.SharedEQ(true))
}

func (r *MemoryNodesRepository) GetAllByUser(ctx context.Context, userID int) ([]*ent.MemoryNode, error) {
	return r.client.MemoryNode.Query().Where(accessibleMemoryNode(userID)).All(ctx)
}

func (r *MemoryNodesRepository) Get(ctx context.Context, id int) (*ent.MemoryNode, error) {
	return r.client.MemoryNode.Get(ctx, id)
}

// GetForUser returns a node the user owns or that is shared.
func (r *MemoryNodesRepository) GetForUser(ctx context.Context, id, userID int) (*ent.MemoryNode, error) {
	return r.client.MemoryNode.Query().
		Where(memorynode.IDEQ(id), accessibleMemoryNode(userID)).
		Only(ctx)
}

// GetOwnedForUser returns a node only if the user owns it.
func (r *MemoryNodesRepository) GetOwnedForUser(ctx context.Context, id, userID int) (*ent.MemoryNode, error) {
	return r.client.MemoryNode.Query().
		Where(memorynode.IDEQ(id), memorynode.UserIDEQ(userID)).
		Only(ctx)
}

func (r *MemoryNodesRepository) GetByIDsForUser(ctx context.Context, ids []int, userID int) ([]*ent.MemoryNode, error) {
	if len(ids) == 0 {
		return []*ent.MemoryNode{}, nil
	}
	return r.client.MemoryNode.Query().
		Where(memorynode.IDIn(ids...), accessibleMemoryNode(userID)).
		All(ctx)
}

func (r *MemoryNodesRepository) Create(
	ctx context.Context,
	name string,
	parents, children, cards []int,
	aliases []string,
	shared bool,
	userID int,
) (*ent.MemoryNode, error) {
	if parents == nil {
		parents = []int{}
	}
	if children == nil {
		children = []int{}
	}
	if cards == nil {
		cards = []int{}
	}
	if aliases == nil {
		aliases = []string{}
	}
	return r.client.MemoryNode.Create().
		SetName(name).
		SetParents(parents).
		SetChildren(children).
		SetCards(cards).
		SetAliases(aliases).
		SetPriorities([]schema.CardsPriority{}).
		SetGroups([]schema.CardsGroup{}).
		SetShared(shared).
		SetUserID(userID).
		Save(ctx)
}

func (r *MemoryNodesRepository) Update(ctx context.Context, partial models.MemoryNodePartial) (*ent.MemoryNode, error) {
	upd := r.client.MemoryNode.UpdateOneID(partial.ID)
	if partial.Name != nil {
		upd = upd.SetName(*partial.Name)
	}
	if partial.Children != nil {
		upd = upd.SetChildren(*partial.Children)
	}
	if partial.Parents != nil {
		upd = upd.SetParents(*partial.Parents)
	}
	if partial.Cards != nil {
		upd = upd.SetCards(*partial.Cards)
	}
	if partial.Aliases != nil {
		upd = upd.SetAliases(*partial.Aliases)
	}
	if partial.Priorities != nil {
		upd = upd.SetPriorities(*partial.Priorities)
	}
	if partial.Groups != nil {
		upd = upd.SetGroups(*partial.Groups)
	}
	if partial.Shared != nil {
		upd = upd.SetShared(*partial.Shared)
	}
	return upd.Save(ctx)
}

func (r *MemoryNodesRepository) Delete(ctx context.Context, id int) error {
	return r.client.MemoryNode.DeleteOneID(id).Exec(ctx)
}

func (r *MemoryNodesRepository) FindByAliasForUser(ctx context.Context, alias string, userID int) (*ent.MemoryNode, error) {
	nodes, err := r.client.MemoryNode.Query().Where(accessibleMemoryNode(userID)).All(ctx)
	if err != nil {
		return nil, err
	}
	target := strings.ToLower(strings.TrimSpace(alias))
	for _, node := range nodes {
		for _, a := range node.Aliases {
			if strings.ToLower(strings.TrimSpace(a)) == target {
				return node, nil
			}
		}
	}
	return nil, ErrNotFound
}
