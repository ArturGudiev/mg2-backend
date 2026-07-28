package models

import "arturgudiev/memoryguard/ent/schema"

// CardItemFull is the API representation of a card item.
type CardItemFull struct {
	ID        int                 `json:"id"`
	Type      schema.CardItemType `json:"type"`
	Text      *string             `json:"text,omitempty"`
	Index     *int                `json:"index,omitempty"`
	Code      *string             `json:"code,omitempty"`
	Extension *string             `json:"extension,omitempty"`
	Formula   *string             `json:"formula,omitempty"`
	ImagePath *string             `json:"imagePath,omitempty"`
	Width     *string             `json:"width,omitempty"`
	Shared    bool                `json:"shared"`
	UserID    int                 `json:"userId"`
}

// CardItemShort is used when creating a card item.
type CardItemShort struct {
	Type      schema.CardItemType `json:"type" binding:"required"`
	Text      *string             `json:"text"`
	Index     *int                `json:"index"`
	Code      *string             `json:"code"`
	Extension *string             `json:"extension"`
	Formula   *string             `json:"formula"`
	ImagePath *string             `json:"imagePath"`
	Width     *string             `json:"width"`
	Shared    *bool               `json:"shared"`
}

// CardItemPartial is used when partially updating a card item.
type CardItemPartial struct {
	ID        int      `json:"id" binding:"required"`
	Text      *string  `json:"text"`
	Index     *int     `json:"index"`
	Code      *string  `json:"code"`
	Extension *string  `json:"extension"`
	Formula   *string  `json:"formula"`
	ImagePath *string  `json:"imagePath"`
	Width     *string  `json:"width"`
	Shared    *bool    `json:"shared"`
}
