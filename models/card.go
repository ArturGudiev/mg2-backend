package models

import "arturgudiev/memoryguard/ent/schema"

// CardFull is the API representation of a card with resolved card items.
type CardFull struct {
	ID            int              `json:"id"`
	Question      []CardItemFull   `json:"question"`
	Answer        []CardItemFull   `json:"answer"`
	QuestionIDs   []int            `json:"questionIds"`
	AnswerIDs     []int            `json:"answerIds"`
	ParentNodes   []int            `json:"parentNodes"`
	Used          int              `json:"used"`
	Needed        int              `json:"needed"`
	Count         int              `json:"count"`
	ReverseCount  int              `json:"reverseCount"`
	PracticeCount int              `json:"practiceCount"`
	UsageType     schema.UsageType `json:"usageType"`
	Shared        bool             `json:"shared"`
	UserID        int              `json:"userId"`
}

// CardShort is used when creating a card.
// Question/Answer may be provided as new item payloads (created) or as existing IDs.
type CardShort struct {
	QuestionItems []CardItemShort   `json:"question"`
	AnswerItems   []CardItemShort   `json:"answer"`
	QuestionIDs   []int             `json:"questionIds"`
	AnswerIDs     []int             `json:"answerIds"`
	ParentNodes   []int             `json:"parentNodes"`
	UsageType     *schema.UsageType `json:"usageType"`
	Shared        *bool             `json:"shared"`
	MemoryNodeID  *int              `json:"memoryNodeId"`
}

// CardPartial is used when updating a card.
type CardPartial struct {
	ID            int               `json:"id" binding:"required"`
	QuestionIDs   *[]int            `json:"questionIds"`
	AnswerIDs     *[]int            `json:"answerIds"`
	QuestionItems *[]CardItemShort  `json:"question"`
	AnswerItems   *[]CardItemShort  `json:"answer"`
	ParentNodes   *[]int            `json:"parentNodes"`
	Used          *int              `json:"used"`
	Needed        *int              `json:"needed"`
	Count         *int              `json:"count"`
	ReverseCount  *int              `json:"reverseCount"`
	PracticeCount *int              `json:"practiceCount"`
	UsageType     *schema.UsageType `json:"usageType"`
	Shared        *bool             `json:"shared"`
}

// NewCardRequest creates a card under a memory node (legacy-compatible shape).
type NewCardRequest struct {
	MemoryNodeID int             `json:"_id" binding:"required"`
	Question     []CardItemShort `json:"question" binding:"required"`
	Answer       []CardItemShort `json:"answer" binding:"required"`
	Shared       *bool           `json:"shared"`
}

// UpdateCardsFieldRequest bulk-updates a numeric field on multiple cards.
type UpdateCardsFieldRequest struct {
	Cards []CardFieldUpdate `json:"cards" binding:"required"`
	Field string            `json:"field" binding:"required"` // count | practiceCount
}

// CardFieldUpdate carries the ID and new field value for bulk updates.
type CardFieldUpdate struct {
	ID            int `json:"id"`
	IDAlt         int `json:"_id"`
	Count         int `json:"count"`
	PracticeCount int `json:"practiceCount"`
}

func (u CardFieldUpdate) ResolvedID() int {
	if u.ID > 0 {
		return u.ID
	}
	return u.IDAlt
}

// CardsByQueryRequest filters cards belonging to a memory node.
type CardsByQueryRequest struct {
	ID       int                   `json:"id" binding:"required"`
	Query    string                `json:"query"`
	Priority *schema.CardsPriority `json:"priority"`
	Group    *schema.CardsGroup    `json:"group"`
}
