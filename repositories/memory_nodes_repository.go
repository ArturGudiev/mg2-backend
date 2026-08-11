package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/memorynode"
	"arturgudiev/memoryguard/ent/memorynodeuser"
	"arturgudiev/memoryguard/ent/predicate"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"context"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

// ErrAccessDenied is returned when a resource exists but the user may not access it.
var ErrAccessDenied = errors.New("access denied")

type MemoryNodesRepository struct {
	client *ent.Client
}

func NewMemoryNodesRepository(client *ent.Client) *MemoryNodesRepository {
	return &MemoryNodesRepository{client: client}
}

// accessibleMemoryNode: owner, or shared node with an explicit memory_node_users row.
func accessibleMemoryNode(userID int) predicate.MemoryNode {
	return memorynode.Or(
		memorynode.UserIDEQ(userID),
		memorynode.And(
			memorynode.SharedEQ(true),
			memorynode.HasMemoryNodeUsersWith(memorynodeuser.UserIDEQ(userID)),
		),
	)
}

func (r *MemoryNodesRepository) GetAllByUser(ctx context.Context, userID int) ([]*ent.MemoryNode, error) {
	return r.client.MemoryNode.Query().Where(accessibleMemoryNode(userID)).All(ctx)
}

// GetRootsByUser returns accessible nodes that have no parents.
func (r *MemoryNodesRepository) GetRootsByUser(ctx context.Context, userID int) ([]*ent.MemoryNode, error) {
	nodes, err := r.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	roots := make([]*ent.MemoryNode, 0)
	for _, n := range nodes {
		if len(n.Parents) == 0 {
			roots = append(roots, n)
		}
	}
	return roots, nil
}

func (r *MemoryNodesRepository) Get(ctx context.Context, id int) (*ent.MemoryNode, error) {
	return r.client.MemoryNode.Get(ctx, id)
}

// GetForUser returns a node the user owns or a shared node they were granted.
func (r *MemoryNodesRepository) GetForUser(ctx context.Context, id, userID int) (*ent.MemoryNode, error) {
	return r.client.MemoryNode.Query().
		Where(memorynode.IDEQ(id), accessibleMemoryNode(userID)).
		Only(ctx)
}

// UserCanAccess reports whether the user owns the node or has an explicit shared grant.
func (r *MemoryNodesRepository) UserCanAccess(ctx context.Context, id, userID int) (bool, error) {
	return r.client.MemoryNode.Query().
		Where(memorynode.IDEQ(id), accessibleMemoryNode(userID)).
		Exist(ctx)
}

// FilterAccessibleIDs returns the subset of ids the user can access, preserving order.
func (r *MemoryNodesRepository) FilterAccessibleIDs(ctx context.Context, ids []int, userID int) ([]int, error) {
	if len(ids) == 0 {
		return []int{}, nil
	}
	nodes, err := r.client.MemoryNode.Query().
		Where(memorynode.IDIn(ids...), accessibleMemoryNode(userID)).
		Select(memorynode.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int]struct{}, len(nodes))
	for _, n := range nodes {
		allowed[n.ID] = struct{}{}
	}
	out := make([]int, 0, len(nodes))
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
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
	description string,
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
		SetDescription(description).
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
	if partial.Description != nil {
		upd = upd.SetDescription(*partial.Description)
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
