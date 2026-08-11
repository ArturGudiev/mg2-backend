package services

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/repositories"
	"context"
	"fmt"
)

type CardItemsService struct {
	repo *repositories.CardItemsRepository
}

func NewCardItemsService(repo *repositories.CardItemsRepository) *CardItemsService {
	return &CardItemsService{repo: repo}
}

func (s *CardItemsService) toFull(item *ent.CardItem) models.CardItemFull {
	return models.CardItemFull{
		ID:        item.ID,
		Type:      item.Type,
		Text:      item.Text,
		Index:     item.Index,
		Code:      item.Code,
		Extension: item.Extension,
		Formula:   item.Formula,
		ImagePath: item.ImagePath,
		Width:     item.Width,
		Shared:    item.Shared,
		UserID:    item.UserID,
	}
}

func (s *CardItemsService) Get(ctx context.Context, id, userID int) (*models.CardItemFull, error) {
	item, err := s.repo.GetForUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	full := s.toFull(item)
	return &full, nil
}

func (s *CardItemsService) GetByIDs(ctx context.Context, ids []int, userID int) ([]models.CardItemFull, error) {
	items, err := s.repo.GetByIDsForUser(ctx, ids, userID)
	if err != nil {
		return nil, err
	}
	result := make([]models.CardItemFull, 0, len(items))
	for _, item := range items {
		result = append(result, s.toFull(item))
	}
	return result, nil
}

func (s *CardItemsService) GetAll(ctx context.Context, userID int) ([]models.CardItemFull, error) {
	items, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]models.CardItemFull, 0, len(items))
	for _, item := range items {
		result = append(result, s.toFull(item))
	}
	return result, nil
}

func (s *CardItemsService) Create(ctx context.Context, short models.CardItemShort, userID int) (*models.CardItemFull, error) {
	if short.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	item, err := s.repo.Create(ctx, short, userID)
	if err != nil {
		return nil, err
	}
	full := s.toFull(item)
	return &full, nil
}

func (s *CardItemsService) CreateMany(ctx context.Context, shorts []models.CardItemShort, userID int) ([]models.CardItemFull, error) {
	items, err := s.repo.CreateMany(ctx, shorts, userID)
	if err != nil {
		return nil, err
	}
	result := make([]models.CardItemFull, 0, len(items))
	for _, item := range items {
		result = append(result, s.toFull(item))
	}
	return result, nil
}

func (s *CardItemsService) Update(ctx context.Context, partial models.CardItemPartial, userID int) (*models.CardItemFull, error) {
	if _, err := s.repo.GetOwnedForUser(ctx, partial.ID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.Update(ctx, partial)
	if err != nil {
		return nil, err
	}
	full := s.toFull(item)
	return &full, nil
}

func (s *CardItemsService) Delete(ctx context.Context, id, userID int) error {
	if _, err := s.repo.GetOwnedForUser(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *CardItemsService) DeleteByIDs(ctx context.Context, ids []int) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// EnsureShared marks card items as shared so grantees can resolve question/answer content.
func (s *CardItemsService) EnsureShared(ctx context.Context, ids []int) error {
	return s.repo.EnsureShared(ctx, ids)
}
