package services

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/repositories"
	"context"
	"fmt"
	"strconv"
	"strings"
)

type CardsService struct {
	cardsRepo      *repositories.CardsRepository
	cardItemsSvc   *CardItemsService
	memoryNodesSvc *MemoryNodesService
}

func NewCardsService(
	cardsRepo *repositories.CardsRepository,
	cardItemsSvc *CardItemsService,
	memoryNodesSvc *MemoryNodesService,
) *CardsService {
	return &CardsService{
		cardsRepo:      cardsRepo,
		cardItemsSvc:   cardItemsSvc,
		memoryNodesSvc: memoryNodesSvc,
	}
}

func (s *CardsService) toFull(ctx context.Context, c *ent.Card, userID int) (*models.CardFull, error) {
	question, err := s.cardItemsSvc.GetByIDs(ctx, c.Question, userID)
	if err != nil {
		return nil, err
	}
	answer, err := s.cardItemsSvc.GetByIDs(ctx, c.Answer, userID)
	if err != nil {
		return nil, err
	}
	questionIDs := c.Question
	if questionIDs == nil {
		questionIDs = []int{}
	}
	answerIDs := c.Answer
	if answerIDs == nil {
		answerIDs = []int{}
	}
	parentNodes := c.ParentNodes
	if parentNodes == nil {
		parentNodes = []int{}
	}
	if question == nil {
		question = []models.CardItemFull{}
	}
	if answer == nil {
		answer = []models.CardItemFull{}
	}
	return &models.CardFull{
		ID:            c.ID,
		Question:      question,
		Answer:        answer,
		QuestionIDs:   questionIDs,
		AnswerIDs:     answerIDs,
		ParentNodes:   parentNodes,
		Used:          c.Used,
		Needed:        c.Needed,
		Count:        c.Count,
		ReverseCount: c.ReverseCount,
		UsageType:    c.UsageType,
		Shared:        c.Shared,
		UserID:        c.UserID,
	}, nil
}

func (s *CardsService) Get(ctx context.Context, id, userID int) (*models.CardFull, error) {
	c, err := s.cardsRepo.GetForUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return s.toFull(ctx, c, userID)
}

func (s *CardsService) GetByIDs(ctx context.Context, ids []int, userID int) ([]*models.CardFull, error) {
	cards, err := s.cardsRepo.GetByIDsForUser(ctx, ids, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.CardFull, 0, len(cards))
	for _, c := range cards {
		full, err := s.toFull(ctx, c, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, full)
	}
	return result, nil
}

func (s *CardsService) resolveItemIDs(ctx context.Context, existing []int, shorts []models.CardItemShort, userID int, shared bool) ([]int, error) {
	ids := append([]int{}, existing...)
	if len(shorts) > 0 {
		prepared := make([]models.CardItemShort, len(shorts))
		for i, short := range shorts {
			prepared[i] = short
			if prepared[i].Shared == nil {
				sharedCopy := shared
				prepared[i].Shared = &sharedCopy
			}
		}
		created, err := s.cardItemsSvc.CreateMany(ctx, prepared, userID)
		if err != nil {
			return nil, err
		}
		for _, item := range created {
			ids = append(ids, item.ID)
		}
	}
	return ids, nil
}

func (s *CardsService) Create(ctx context.Context, short models.CardShort, userID int) (*models.CardFull, error) {
	parentNodes := short.ParentNodes
	if parentNodes == nil {
		parentNodes = []int{}
	}
	if short.MemoryNodeID != nil {
		parentNodes = append(parentNodes, *short.MemoryNodeID)
	}

	shared := false
	if short.Shared != nil {
		shared = *short.Shared
	} else if short.MemoryNodeID != nil {
		if node, err := s.memoryNodesSvc.Get(ctx, *short.MemoryNodeID, userID); err == nil && node.Shared {
			shared = true
		}
	}

	questionIDs, err := s.resolveItemIDs(ctx, short.QuestionIDs, short.QuestionItems, userID, shared)
	if err != nil {
		return nil, err
	}
	answerIDs, err := s.resolveItemIDs(ctx, short.AnswerIDs, short.AnswerItems, userID, shared)
	if err != nil {
		return nil, err
	}

	usageType := schema.UsageTypeCommon
	if short.UsageType != nil {
		usageType = *short.UsageType
	}

	created, err := s.cardsRepo.Create(ctx, questionIDs, answerIDs, parentNodes, usageType, shared, userID)
	if err != nil {
		return nil, err
	}

	for _, nodeID := range parentNodes {
		_ = s.memoryNodesSvc.AddCardID(ctx, nodeID, created.ID)
	}

	return s.Get(ctx, created.ID, userID)
}

func (s *CardsService) CreateUnderNode(ctx context.Context, req models.NewCardRequest, userID int) (*models.CardFull, error) {
	return s.Create(ctx, models.CardShort{
		QuestionItems: req.Question,
		AnswerItems:   req.Answer,
		MemoryNodeID:  &req.MemoryNodeID,
		Shared:        req.Shared,
	}, userID)
}

func (s *CardsService) Update(ctx context.Context, partial models.CardPartial, userID int) (*models.CardFull, error) {
	existing, err := s.cardsRepo.GetOwnedForUser(ctx, partial.ID, userID)
	if err != nil {
		return nil, err
	}
	shared := existing.Shared
	if partial.Shared != nil {
		shared = *partial.Shared
	}
	if partial.QuestionItems != nil {
		ids, err := s.resolveItemIDs(ctx, nil, *partial.QuestionItems, userID, shared)
		if err != nil {
			return nil, err
		}
		partial.QuestionIDs = &ids
	}
	if partial.AnswerItems != nil {
		ids, err := s.resolveItemIDs(ctx, nil, *partial.AnswerItems, userID, shared)
		if err != nil {
			return nil, err
		}
		partial.AnswerIDs = &ids
	}
	if _, err := s.cardsRepo.Update(ctx, partial); err != nil {
		return nil, err
	}
	if partial.Shared != nil {
		sharedVal := *partial.Shared
		itemIDs := append(append([]int{}, existing.Question...), existing.Answer...)
		for _, itemID := range itemIDs {
			_, _ = s.cardItemsSvc.Update(ctx, models.CardItemPartial{ID: itemID, Shared: &sharedVal}, userID)
		}
	}
	return s.Get(ctx, partial.ID, userID)
}

func (s *CardsService) Delete(ctx context.Context, id, userID int) error {
	c, err := s.cardsRepo.GetOwnedForUser(ctx, id, userID)
	if err != nil {
		return err
	}
	for _, nodeID := range c.ParentNodes {
		_ = s.memoryNodesSvc.RemoveCardID(ctx, nodeID, id)
	}
	itemIDs := append(append([]int{}, c.Question...), c.Answer...)
	_ = s.cardItemsSvc.DeleteByIDs(ctx, itemIDs)
	return s.cardsRepo.Delete(ctx, id)
}

func (s *CardsService) DeleteMany(ctx context.Context, ids []int, userID int) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id, userID); err != nil && !ent.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (s *CardsService) UpdateField(ctx context.Context, req models.UpdateCardsFieldRequest, userID int) error {
	field := strings.TrimSpace(req.Field)
	for _, item := range req.Cards {
		partial := models.CardPartial{ID: item.ResolvedID()}
		if partial.ID <= 0 {
			return fmt.Errorf("card id is required")
		}
		if _, err := s.cardsRepo.GetForUser(ctx, partial.ID, userID); err != nil {
			return err
		}
		switch field {
		case "count":
			count := item.Count
			partial.Count = &count
		default:
			return fmt.Errorf("unsupported field: %s", field)
		}
		if _, err := s.cardsRepo.Update(ctx, partial); err != nil {
			return err
		}
	}
	return nil
}

func (s *CardsService) GetByMemoryNodeID(ctx context.Context, nodeID, userID int) ([]*models.CardFull, error) {
	node, err := s.memoryNodesSvc.Get(ctx, nodeID, userID)
	if err != nil {
		return nil, err
	}
	return s.GetByIDs(ctx, node.Cards, userID)
}

func (s *CardsService) CardsByQuery(ctx context.Context, req models.CardsByQueryRequest, userID int) ([]*models.CardFull, error) {
	var ids []int
	switch {
	case req.Priority != nil:
		ids = req.Priority.Cards
	case req.Group != nil:
		ids = req.Group.Cards
	default:
		node, err := s.memoryNodesSvc.Get(ctx, req.ID, userID)
		if err != nil {
			return nil, err
		}
		ids = node.Cards
	}
	cards, err := s.GetByIDs(ctx, ids, userID)
	if err != nil {
		return nil, err
	}
	return filterCards(cards, req.Query), nil
}

func (s *CardsService) IncrementCount(ctx context.Context, id, userID int) (*models.CardFull, error) {
	if _, err := s.cardsRepo.GetForUser(ctx, id, userID); err != nil {
		return nil, err
	}
	if _, err := s.cardsRepo.IncrementCount(ctx, id); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, userID)
}

func (s *CardsService) DecrementCount(ctx context.Context, id, userID int) (*models.CardFull, error) {
	if _, err := s.cardsRepo.GetForUser(ctx, id, userID); err != nil {
		return nil, err
	}
	if _, err := s.cardsRepo.DecrementCount(ctx, id); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, userID)
}

// filterCards applies a simple query language similar to the old selectCards helper.
// Supported: "count < N", "count <= N", "count > N", "count >= N", "count = N", "count == N",
// "limit N", "--limit N". Quiz tokens like "quiz", "pquiz", and "-until N" are ignored.
func filterCards(cards []*models.CardFull, query string) []*models.CardFull {
	query = strings.TrimSpace(query)
	if query == "" {
		return cards
	}

	result := cards
	remaining := query

	for remaining != "" {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		if m := matchPrefix(remaining, "--limit"); m != "" {
			n, rest, ok := parseIntThenRest(m)
			if ok {
				if n < len(result) {
					result = result[:n]
				}
				remaining = rest
				continue
			}
		}

		if m := matchPrefix(remaining, "limit"); m != "" {
			n, rest, ok := parseIntThenRest(m)
			if ok {
				if n < len(result) {
					result = result[:n]
				}
				remaining = rest
				continue
			}
		}

		ops := []string{"<=", ">=", "===", "==", "=", "<", ">"}
		matched := false
		for _, op := range ops {
			subject, value, rest, ok := parseComparison(remaining, op)
			if !ok {
				continue
			}
			result = applyComparison(result, subject, op, value)
			remaining = rest
			matched = true
			break
		}
		if !matched {
			// Skip quiz / unknown tokens so mixed queries still filter.
			parts := strings.Fields(remaining)
			if len(parts) == 0 {
				break
			}
			remaining = strings.TrimSpace(remaining[len(parts[0]):])
		}
	}
	return result
}

func matchPrefix(s, prefix string) string {
	if strings.HasPrefix(strings.ToLower(s), prefix) {
		return strings.TrimSpace(s[len(prefix):])
	}
	return ""
}

func parseIntThenRest(s string) (int, string, bool) {
	s = strings.TrimSpace(s)
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0, s, false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, s, false
	}
	rest := strings.TrimSpace(s[len(parts[0]):])
	return n, rest, true
}

func parseComparison(s, op string) (subject string, value int, rest string, ok bool) {
	idx := strings.Index(s, op)
	if idx < 0 {
		return "", 0, s, false
	}
	subject = strings.TrimSpace(s[:idx])
	if subject == "" || strings.Contains(subject, " ") {
		return "", 0, s, false
	}
	after := strings.TrimSpace(s[idx+len(op):])
	value, rest, ok = parseIntThenRest(after)
	return subject, value, rest, ok
}

func applyComparison(cards []*models.CardFull, subject, op string, value int) []*models.CardFull {
	out := make([]*models.CardFull, 0, len(cards))
	for _, c := range cards {
		var field int
		switch strings.ToLower(subject) {
		case "count":
			field = c.Count
		case "used":
			field = c.Used
		case "needed":
			field = c.Needed
		default:
			continue
		}
		keep := false
		switch op {
		case "<":
			keep = field < value
		case "<=":
			keep = field <= value
		case ">":
			keep = field > value
		case ">=":
			keep = field >= value
		case "=", "==", "===":
			keep = field == value
		}
		if keep {
			out = append(out, c)
		}
	}
	return out
}
