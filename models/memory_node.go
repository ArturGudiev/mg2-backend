package models

import "arturgudiev/memoryguard/ent/schema"

// MemoryNodeFull is the API representation of a memory node.
type MemoryNodeFull struct {
	ID         int                    `json:"id"`
	Name       string                 `json:"name"`
	Children   []int                  `json:"children"`
	Parents    []int                  `json:"parents"`
	Cards      []int                  `json:"cards"`
	Aliases    []string               `json:"aliases"`
	Priorities []schema.CardsPriority `json:"priorities"`
	Groups     []schema.CardsGroup    `json:"groups"`
	Shared     bool                   `json:"shared"`
	UserID     int                    `json:"userId"`
}

// MemoryNodeShort is used when creating a memory node.
type MemoryNodeShort struct {
	Name     string   `json:"name" binding:"required"`
	Parents  []int    `json:"parents"`
	Aliases  []string `json:"aliases"`
	Shared   bool     `json:"shared"`
	UserID   int      `json:"userId"`
	Children []int    `json:"children"`
	Cards    []int    `json:"cards"`
}

// MemoryNodePartial is used when partially updating a memory node.
type MemoryNodePartial struct {
	ID         int                     `json:"id" binding:"required"`
	Name       *string                 `json:"name"`
	Children   *[]int                  `json:"children"`
	Parents    *[]int                  `json:"parents"`
	Cards      *[]int                  `json:"cards"`
	Aliases    *[]string               `json:"aliases"`
	Priorities *[]schema.CardsPriority `json:"priorities"`
	Groups     *[]schema.CardsGroup    `json:"groups"`
	Shared     *bool                   `json:"shared"`
	UserID     *int                    `json:"userId"`
}

// NewMemoryNodeRequest creates a memory node and optionally links it to parents.
type NewMemoryNodeRequest struct {
	MemoryNode MemoryNodeShort `json:"memoryNode" binding:"required"`
}

// MemoryNodePathItem is one step in a parents-path chain (root → leaf).
type MemoryNodePathItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
