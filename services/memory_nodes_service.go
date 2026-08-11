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
	repo                *repositories.MemoryNodesRepository
	memoryNodeUsersRepo *repositories.MemoryNodeUsersRepository
	cardsRepo           *repositories.CardsRepository
	cardUsersRepo       *repositories.CardUsersRepository
	cardUserCountsRepo  *repositories.CardUserCountsRepository
	cardItemsRepo       *repositories.CardItemsRepository
}

func NewMemoryNodesService(
	repo *repositories.MemoryNodesRepository,
	memoryNodeUsersRepo *repositories.MemoryNodeUsersRepository,
	cardsRepo *repositories.CardsRepository,
	cardUsersRepo *repositories.CardUsersRepository,
	cardUserCountsRepo *repositories.CardUserCountsRepository,
	cardItemsRepo *repositories.CardItemsRepository,
) *MemoryNodesService {
	return &MemoryNodesService{
		repo:                repo,
		memoryNodeUsersRepo: memoryNodeUsersRepo,
		cardsRepo:           cardsRepo,
		cardUsersRepo:       cardUsersRepo,
		cardUserCountsRepo:  cardUserCountsRepo,
		cardItemsRepo:       cardItemsRepo,
	}
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
		ID:          node.ID,
		Name:        node.Name,
		Description: node.Description,
		Children:    children,
		Parents:     parents,
		Cards:       cards,
		Aliases:     aliases,
		Priorities:  priorities,
		Groups:      groups,
		Shared:      node.Shared,
		UserID:      node.UserID,
	}
}

// toFullForUser returns a node with cards/children filtered to those the user can access.
func (s *MemoryNodesService) toFullForUser(ctx context.Context, node *ent.MemoryNode, userID int) (*models.MemoryNodeFull, error) {
	full := s.toFull(node)
	if full == nil {
		return nil, nil
	}
	accessibleCards, err := s.cardsRepo.FilterAccessibleIDs(ctx, full.Cards, userID)
	if err != nil {
		return nil, err
	}
	full.Cards = accessibleCards
	accessibleChildren, err := s.repo.FilterAccessibleIDs(ctx, full.Children, userID)
	if err != nil {
		return nil, err
	}
	full.Children = accessibleChildren
	return full, nil
}

func (s *MemoryNodesService) Get(ctx context.Context, id, userID int) (*models.MemoryNodeFull, error) {
	node, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.repo.UserCanAccess(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, repositories.ErrAccessDenied
	}
	return s.toFullForUser(ctx, node, userID)
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
			full, err := s.toFullForUser(ctx, n, userID)
			if err != nil {
				return nil, err
			}
			result = append(result, full)
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
		full, err := s.toFullForUser(ctx, n, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, full)
	}
	return result, nil
}

func (s *MemoryNodesService) GetRoots(ctx context.Context, userID int) ([]*models.MemoryNodeFull, error) {
	nodes, err := s.repo.GetRootsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.MemoryNodeFull, 0, len(nodes))
	for _, n := range nodes {
		full, err := s.toFullForUser(ctx, n, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, full)
	}
	return result, nil
}

func (s *MemoryNodesService) Create(ctx context.Context, short models.MemoryNodeShort, userID int) (*models.MemoryNodeFull, error) {
	if short.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	short.UserID = userID
	node, err := s.repo.Create(ctx, short.Name, short.Description, short.Parents, short.Children, short.Cards, short.Aliases, short.Shared, short.UserID)
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
	node, err := s.repo.GetOwnedForUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if node.Shared {
		return s.deleteSharedCascade(ctx, node, userID, map[int]struct{}{})
	}
	if err := s.memoryNodeUsersRepo.DeleteByMemoryNodeID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// deleteSharedCascade deletes a shared node, its owned descendants, cards, and all grant links.
func (s *MemoryNodesService) deleteSharedCascade(
	ctx context.Context,
	node *ent.MemoryNode,
	userID int,
	seen map[int]struct{},
) error {
	if node == nil {
		return nil
	}
	if _, ok := seen[node.ID]; ok {
		return nil
	}
	seen[node.ID] = struct{}{}

	for _, childID := range node.Children {
		child, err := s.repo.GetOwnedForUser(ctx, childID, userID)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return err
		}
		if err := s.deleteSharedCascade(ctx, child, userID, seen); err != nil {
			return err
		}
	}

	for _, cardID := range node.Cards {
		if err := s.deleteCardFully(ctx, cardID); err != nil {
			return err
		}
	}

	if err := s.memoryNodeUsersRepo.DeleteByMemoryNodeID(ctx, node.ID); err != nil {
		return err
	}
	if err := s.unlinkFromParents(ctx, node); err != nil {
		return err
	}
	return s.repo.Delete(ctx, node.ID)
}

func (s *MemoryNodesService) unlinkFromParents(ctx context.Context, node *ent.MemoryNode) error {
	for _, parentID := range node.Parents {
		parent, err := s.repo.Get(ctx, parentID)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return err
		}
		filtered := make([]int, 0, len(parent.Children))
		for _, id := range parent.Children {
			if id != node.ID {
				filtered = append(filtered, id)
			}
		}
		if _, err := s.repo.Update(ctx, models.MemoryNodePartial{ID: parentID, Children: &filtered}); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryNodesService) deleteCardFully(ctx context.Context, cardID int) error {
	card, err := s.cardsRepo.Get(ctx, cardID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	for _, nodeID := range card.ParentNodes {
		_ = s.RemoveCardID(ctx, nodeID, cardID)
	}
	if err := s.cardUsersRepo.DeleteByCardID(ctx, cardID); err != nil {
		return err
	}
	if err := s.cardUserCountsRepo.DeleteByCardID(ctx, cardID); err != nil {
		return err
	}
	itemIDs := append(append([]int{}, card.Question...), card.Answer...)
	if err := s.cardItemsRepo.DeleteByIDs(ctx, itemIDs); err != nil {
		return err
	}
	return s.cardsRepo.Delete(ctx, cardID)
}

func (s *MemoryNodesService) GetByAlias(ctx context.Context, alias string, userID int) (*models.MemoryNodeFull, error) {
	node, err := s.repo.FindByAliasForUser(ctx, alias, userID)
	if err != nil {
		return nil, err
	}
	return s.toFullForUser(ctx, node, userID)
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

// ListGrantedUserIDs returns users who have an explicit memory_node_users grant.
func (s *MemoryNodesService) ListGrantedUserIDs(ctx context.Context, nodeID int) ([]int, error) {
	return s.memoryNodeUsersRepo.ListUserIDs(ctx, nodeID)
}

// GrantAccess gives a user access to a shared memory node (memory_node_users row).
// Caller must cascade card grants separately if needed.
func (s *MemoryNodesService) GrantAccess(ctx context.Context, nodeID, granteeUserID, actorUserID int) error {
	node, err := s.repo.GetOwnedForUser(ctx, nodeID, actorUserID)
	if err != nil {
		return err
	}
	if !node.Shared {
		return fmt.Errorf("memory node must be shared before granting access")
	}
	if granteeUserID <= 0 {
		return fmt.Errorf("user id is required")
	}
	return s.memoryNodeUsersRepo.EnsureLink(ctx, nodeID, granteeUserID)
}

// GetByID returns a memory node by ID without access checks (admin / internal use).
func (s *MemoryNodesService) GetByID(ctx context.Context, id int) (*models.MemoryNodeFull, error) {
	node, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toFull(node), nil
}

// EnsureUserLink creates a memory_node_users row for an already-verified shared node.
func (s *MemoryNodesService) EnsureUserLink(ctx context.Context, nodeID, granteeUserID int) error {
	if granteeUserID <= 0 {
		return fmt.Errorf("user id is required")
	}
	return s.memoryNodeUsersRepo.EnsureLink(ctx, nodeID, granteeUserID)
}

// RemoveUserLink deletes the memory_node_users row for a user on a node (no-op if missing).
func (s *MemoryNodesService) RemoveUserLink(ctx context.Context, nodeID, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user id is required")
	}
	return s.memoryNodeUsersRepo.DeleteLink(ctx, nodeID, userID)
}
