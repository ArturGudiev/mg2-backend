package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/carditem"
	"arturgudiev/memoryguard/ent/predicate"
	"arturgudiev/memoryguard/models"
	"context"
)

type CardItemsRepository struct {
	client *ent.Client
}

func NewCardItemsRepository(client *ent.Client) *CardItemsRepository {
	return &CardItemsRepository{client: client}
}

func accessibleCardItem(userID int) predicate.CardItem {
	return carditem.Or(carditem.UserIDEQ(userID), carditem.SharedEQ(true))
}

func (r *CardItemsRepository) Get(ctx context.Context, id int) (*ent.CardItem, error) {
	return r.client.CardItem.Get(ctx, id)
}

// GetForUser returns a card item the user owns or that is shared.
func (r *CardItemsRepository) GetForUser(ctx context.Context, id, userID int) (*ent.CardItem, error) {
	return r.client.CardItem.Query().
		Where(carditem.IDEQ(id), accessibleCardItem(userID)).
		Only(ctx)
}

// GetOwnedForUser returns a card item only if the user owns it.
func (r *CardItemsRepository) GetOwnedForUser(ctx context.Context, id, userID int) (*ent.CardItem, error) {
	return r.client.CardItem.Query().
		Where(carditem.IDEQ(id), carditem.UserIDEQ(userID)).
		Only(ctx)
}

func (r *CardItemsRepository) GetByIDsForUser(ctx context.Context, ids []int, userID int) ([]*ent.CardItem, error) {
	if len(ids) == 0 {
		return []*ent.CardItem{}, nil
	}
	items, err := r.client.CardItem.Query().
		Where(carditem.IDIn(ids...), accessibleCardItem(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]*ent.CardItem, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	ordered := make([]*ent.CardItem, 0, len(ids))
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			ordered = append(ordered, item)
		}
	}
	return ordered, nil
}

func (r *CardItemsRepository) GetAllByUser(ctx context.Context, userID int) ([]*ent.CardItem, error) {
	return r.client.CardItem.Query().Where(accessibleCardItem(userID)).All(ctx)
}

func (r *CardItemsRepository) Create(ctx context.Context, short models.CardItemShort, userID int) (*ent.CardItem, error) {
	shared := false
	if short.Shared != nil {
		shared = *short.Shared
	}
	builder := r.client.CardItem.Create().
		SetType(short.Type).
		SetShared(shared).
		SetUserID(userID)
	if short.Text != nil {
		builder = builder.SetText(*short.Text)
	}
	if short.Index != nil {
		builder = builder.SetIndex(*short.Index)
	}
	if short.Code != nil {
		builder = builder.SetCode(*short.Code)
	}
	if short.Extension != nil {
		builder = builder.SetExtension(*short.Extension)
	}
	if short.Formula != nil {
		builder = builder.SetFormula(*short.Formula)
	}
	if short.ImagePath != nil {
		builder = builder.SetImagePath(*short.ImagePath)
	}
	if short.Width != nil {
		builder = builder.SetWidth(*short.Width)
	}
	return builder.Save(ctx)
}

func (r *CardItemsRepository) CreateMany(ctx context.Context, shorts []models.CardItemShort, userID int) ([]*ent.CardItem, error) {
	result := make([]*ent.CardItem, 0, len(shorts))
	for _, short := range shorts {
		item, err := r.Create(ctx, short, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *CardItemsRepository) Update(ctx context.Context, partial models.CardItemPartial) (*ent.CardItem, error) {
	upd := r.client.CardItem.UpdateOneID(partial.ID)
	if partial.Text != nil {
		upd = upd.SetText(*partial.Text)
	}
	if partial.Index != nil {
		upd = upd.SetIndex(*partial.Index)
	}
	if partial.Code != nil {
		upd = upd.SetCode(*partial.Code)
	}
	if partial.Extension != nil {
		upd = upd.SetExtension(*partial.Extension)
	}
	if partial.Formula != nil {
		upd = upd.SetFormula(*partial.Formula)
	}
	if partial.ImagePath != nil {
		upd = upd.SetImagePath(*partial.ImagePath)
	}
	if partial.Width != nil {
		upd = upd.SetWidth(*partial.Width)
	}
	if partial.Shared != nil {
		upd = upd.SetShared(*partial.Shared)
	}
	return upd.Save(ctx)
}

func (r *CardItemsRepository) Delete(ctx context.Context, id int) error {
	return r.client.CardItem.DeleteOneID(id).Exec(ctx)
}

func (r *CardItemsRepository) DeleteByIDs(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.CardItem.Delete().Where(carditem.IDIn(ids...)).Exec(ctx)
	return err
}
