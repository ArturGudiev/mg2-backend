package handlers

// IDsRequest represents a request with a list of IDs.
type IDsRequest struct {
	IDs []int `json:"ids" binding:"required"`
}
