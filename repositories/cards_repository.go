package repositories

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/card"
	"arturgudiev/memoryguard/ent/predicate"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"context"
)

type CardsRepository struct {
	client *ent.Client
}

func NewCardsRepository(client *ent.Client) *CardsRepository {
	return &CardsRepository{client: client}
}

func accessibleCard(userID int) predicate.Card {
	return card.Or(card.UserIDEQ(userID), card.SharedEQ(true))
}

func (r *CardsRepository) Get(ctx context.Context, id int) (*ent.Card, error) {
	return r.client.Card.Get(ctx, id)
}

// GetForUser returns a card the user owns or that is shared.
func (r *CardsRepository) GetForUser(ctx context.Context, id, userID int) (*ent.Card, error) {
	return r.client.Card.Query().
		Where(card.IDEQ(id), accessibleCard(userID)).
		Only(ctx)
}

// GetOwnedForUser returns a card only if the user owns it.
func (r *CardsRepository) GetOwnedForUser(ctx context.Context, id, userID int) (*ent.Card, error) {
	return r.client.Card.Query().
		Where(card.IDEQ(id), card.UserIDEQ(userID)).
		Only(ctx)
}

func (r *CardsRepository) GetByIDsForUser(ctx context.Context, ids []int, userID int) ([]*ent.Card, error) {
	if len(ids) == 0 {
		return []*ent.Card{}, nil
	}
	cards, err := r.client.Card.Query().
		Where(card.IDIn(ids...), accessibleCard(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]*ent.Card, len(cards))
	for _, c := range cards {
		byID[c.ID] = c
	}
	ordered := make([]*ent.Card, 0, len(ids))
	for _, id := range ids {
		if c, ok := byID[id]; ok {
			ordered = append(ordered, c)
		}
	}
	return ordered, nil
}

func (r *CardsRepository) Create(
	ctx context.Context,
	questionIDs, answerIDs, parentNodes []int,
	usageType schema.UsageType,
	shared bool,
	userID int,
) (*ent.Card, error) {
	if questionIDs == nil {
		questionIDs = []int{}
	}
	if answerIDs == nil {
		answerIDs = []int{}
	}
	if parentNodes == nil {
		parentNodes = []int{}
	}
	if usageType == "" {
		usageType = schema.UsageTypeCommon
	}
	return r.client.Card.Create().
		SetQuestion(questionIDs).
		SetAnswer(answerIDs).
		SetParentNodes(parentNodes).
		SetUsed(0).
		SetNeeded(0).
		SetCount(0).
		SetReverseCount(0).
		SetPracticeCount(0).
		SetUsageType(usageType).
		SetShared(shared).
		SetUserID(userID).
		Save(ctx)
}

func (r *CardsRepository) Update(ctx context.Context, partial models.CardPartial) (*ent.Card, error) {
	upd := r.client.Card.UpdateOneID(partial.ID)
	if partial.QuestionIDs != nil {
		upd = upd.SetQuestion(*partial.QuestionIDs)
	}
	if partial.AnswerIDs != nil {
		upd = upd.SetAnswer(*partial.AnswerIDs)
	}
	if partial.ParentNodes != nil {
		upd = upd.SetParentNodes(*partial.ParentNodes)
	}
	if partial.Used != nil {
		upd = upd.SetUsed(*partial.Used)
	}
	if partial.Needed != nil {
		upd = upd.SetNeeded(*partial.Needed)
	}
	if partial.Count != nil {
		upd = upd.SetCount(*partial.Count)
	}
	if partial.ReverseCount != nil {
		upd = upd.SetReverseCount(*partial.ReverseCount)
	}
	if partial.PracticeCount != nil {
		upd = upd.SetPracticeCount(*partial.PracticeCount)
	}
	if partial.UsageType != nil {
		upd = upd.SetUsageType(*partial.UsageType)
	}
	if partial.Shared != nil {
		upd = upd.SetShared(*partial.Shared)
	}
	return upd.Save(ctx)
}

func (r *CardsRepository) Delete(ctx context.Context, id int) error {
	return r.client.Card.DeleteOneID(id).Exec(ctx)
}

func (r *CardsRepository) IncrementCount(ctx context.Context, id int) (*ent.Card, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.client.Card.UpdateOneID(id).SetCount(c.Count + 1).Save(ctx)
}

func (r *CardsRepository) DecrementCount(ctx context.Context, id int) (*ent.Card, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	next := c.Count - 1
	if next < 0 {
		next = 0
	}
	return r.client.Card.UpdateOneID(id).SetCount(next).Save(ctx)
}

func (r *CardsRepository) IncrementPracticeCount(ctx context.Context, id int) (*ent.Card, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.client.Card.UpdateOneID(id).SetPracticeCount(c.PracticeCount + 1).Save(ctx)
}

func (r *CardsRepository) DecrementPracticeCount(ctx context.Context, id int) (*ent.Card, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	next := c.PracticeCount - 1
	if next < 0 {
		next = 0
	}
	return r.client.Card.UpdateOneID(id).SetPracticeCount(next).Save(ctx)
}
