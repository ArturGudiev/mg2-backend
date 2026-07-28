package services

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/repositories"
	"context"
	"errors"
	"fmt"
)

type MemoryNodesService struct {
	repo *repositories.MemoryNodesRepository
}

func NewMemoryNodesService(repo *repositories.MemoryNodesRepository) *MemoryNodesService {
	return &MemoryNodesService{repo: repo}
}

func (s *MemoryNodesService) toFull(node *ent.MemoryNode) *models.MemoryNodeFull {
	if node == nil {
		return nil
	}
	priorities := node.Priorities
	if priorities == nil {
		priorities = []schema.CardsPriority{}
	}
	groups := node.Groups
	if groups == nil {
		groups = []schema.CardsGroup{}
	}
	children := node.Children
	if children == nil {
		children = []int{}
	}
	parents := node.Parents
	if parents == nil {
		parents = []int{}
	}
	cards := node.Cards
	if cards == nil {
		cards = []int{}
	}
	aliases := node.Aliases
	if aliases == nil {
		aliases = []string{}
	}
	return &models.MemoryNodeFull{
		ID:         node.ID,
		Name:       node.Name,
		Children:   children,
		Parents:    parents,
		Cards:      cards,
		Aliases:    aliases,
		Priorities: priorities,
		Groups:     groups,
		Shared:     node.Shared,
		UserID:     node.UserID,
	}
}

func (s *MemoryNodesService) Get(ctx context.Context, id, userID int) (*models.MemoryNodeFull, error) {
	node, err := s.repo.GetForUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return s.toFull(node), nil
}

func (s *MemoryNodesService) GetByIDs(ctx context.Context, ids []int, userID int) ([]*models.MemoryNodeFull, error) {
	nodes, err := s.repo.GetByIDsForUser(ctx, ids, userID)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]*ent.MemoryNode, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = n
	}
	result := make([]*models.MemoryNodeFull, 0, len(ids))
	for _, id := range ids {
		if n, ok := byID[id]; ok {
			result = append(result, s.toFull(n))
		}
	}
	return result, nil
}

func (s *MemoryNodesService) GetAll(ctx context.Context, userID int) ([]*models.MemoryNodeFull, error) {
	nodes, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.MemoryNodeFull, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, s.toFull(n))
	}
	return result, nil
}

func (s *MemoryNodesService) Create(ctx context.Context, short models.MemoryNodeShort, userID int) (*models.MemoryNodeFull, error) {
	if short.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	short.UserID = userID
	node, err := s.repo.Create(ctx, short.Name, short.Parents, short.Children, short.Cards, short.Aliases, short.Shared, short.UserID)
	if err != nil {
		return nil, err
	}

	// Link as child only on parents the creator owns.
	for _, parentID := range short.Parents {
		parent, err := s.repo.GetOwnedForUser(ctx, parentID, userID)
		if err != nil {
			continue
		}
		children := append([]int{}, parent.Children...)
		children = append(children, node.ID)
		_, _ = s.repo.Update(ctx, models.MemoryNodePartial{ID: parentID, Children: &children})
	}

	return s.Get(ctx, node.ID, userID)
}

func (s *MemoryNodesService) Update(ctx context.Context, partial models.MemoryNodePartial, userID int) (*models.MemoryNodeFull, error) {
	if _, err := s.repo.GetOwnedForUser(ctx, partial.ID, userID); err != nil {
		return nil, err
	}
	if _, err := s.repo.Update(ctx, partial); err != nil {
		return nil, err
	}
	return s.Get(ctx, partial.ID, userID)
}

func (s *MemoryNodesService) Delete(ctx context.Context, id, userID int) error {
	if _, err := s.repo.GetOwnedForUser(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *MemoryNodesService) GetByAlias(ctx context.Context, alias string, userID int) (*models.MemoryNodeFull, error) {
	node, err := s.repo.FindByAliasForUser(ctx, alias, userID)
	if err != nil {
		return nil, err
	}
	return s.toFull(node), nil
}

// GetParentsPath walks the first-parent chain and returns root → current.
func (s *MemoryNodesService) GetParentsPath(ctx context.Context, id, userID int) ([]models.MemoryNodePathItem, error) {
	node, err := s.repo.GetForUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	chain := []models.MemoryNodePathItem{{ID: node.ID, Name: node.Name}}
	seen := map[int]bool{node.ID: true}
	current := node

	for len(current.Parents) > 0 {
		parentID := current.Parents[0]
		if parentID <= 0 || seen[parentID] {
			break
		}
		parent, err := s.repo.GetForUser(ctx, parentID, userID)
		if err != nil {
			break
		}
		chain = append(chain, models.MemoryNodePathItem{ID: parent.ID, Name: parent.Name})
		seen[parent.ID] = true
		current = parent
	}

	// Reverse leaf→root into root→leaf for display (dashboard-ui style).
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, nil
}

func (s *MemoryNodesService) IsAliasUsed(ctx context.Context, alias string, userID int) (bool, error) {
	_, err := s.repo.FindByAliasForUser(ctx, alias, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MemoryNodesService) AddCardID(ctx context.Context, nodeID, cardID int) error {
	node, err := s.repo.Get(ctx, nodeID)
	if err != nil {
		return err
	}
	cards := append([]int{}, node.Cards...)
	for _, existing := range cards {
		if existing == cardID {
			return nil
		}
	}
	cards = append(cards, cardID)
	_, err = s.repo.Update(ctx, models.MemoryNodePartial{ID: nodeID, Cards: &cards})
	return err
}

func (s *MemoryNodesService) RemoveCardID(ctx context.Context, nodeID, cardID int) error {
	node, err := s.repo.Get(ctx, nodeID)
	if err != nil {
		return err
	}
	filtered := make([]int, 0, len(node.Cards))
	for _, id := range node.Cards {
		if id != cardID {
			filtered = append(filtered, id)
		}
	}
	_, err = s.repo.Update(ctx, models.MemoryNodePartial{ID: nodeID, Cards: &filtered})
	return err
}
