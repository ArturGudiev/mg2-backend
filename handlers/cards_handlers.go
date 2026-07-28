package handlers

import (
	"encoding/json"
	"net/http"

	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/models"

	"github.com/gin-gonic/gin"
)

func jsonNumberArray(raw string, dest *[]int) error {
	return json.Unmarshal([]byte(raw), dest)
}

// GetCardByID handles GET /card/:id
// @Summary      Get card by ID
// @Description  Returns a card with resolved question/answer card items for the authenticated user
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /card/{id} [get]
func (h *Handler) GetCardByID(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	card, err := h.App.CardsService.Get(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// ListCards handles GET /cards?ids=[1,2,3]
// @Summary      List cards by IDs (query)
// @Description  Returns cards for the given JSON ids query parameter owned by the authenticated user
// @Tags         cards
// @Produce      json
// @Param        ids  query     string  true  "JSON array of IDs, e.g. [1,2,3]"
// @Success      200  {array}   models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /cards [get]
func (h *Handler) ListCards(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query required"})
		return
	}
	var ids []int
	if err := jsonNumberArray(idsParam, &ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ids query"})
		return
	}
	cards, err := h.App.CardsService.GetByIDs(c.Request.Context(), ids, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// GetCardsByIDs handles POST /get-cards
// @Summary      Get cards by IDs
// @Description  Returns multiple cards by their IDs for the authenticated user
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of card IDs"
// @Success      200      {array}   models.CardFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /get-cards [post]
func (h *Handler) GetCardsByIDs(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cards, err := h.App.CardsService.GetByIDs(c.Request.Context(), req.IDs, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// NewCard handles POST /new-card
// @Summary      Create card
// @Description  Creates card items and a card under a memory node for the authenticated user
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      models.NewCardRequest  true  "New card request"
// @Success      200      {object}  models.CardFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /new-card [post]
func (h *Handler) NewCard(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.NewCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireAdminForShared(c, req.Shared != nil && *req.Shared) {
		return
	}
	card, err := h.App.CardsService.CreateUnderNode(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// UpdateCard handles PUT /update-card
// @Summary      Update card
// @Description  Partially updates an existing card owned by the authenticated user
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      models.CardPartial  true  "Card update request"
// @Success      200      {object}  models.CardFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /update-card [put]
func (h *Handler) UpdateCard(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.CardPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireAdminForShared(c, req.Shared != nil) {
		return
	}
	card, err := h.App.CardsService.Update(c.Request.Context(), req, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// UpdateCardsField handles POST /update-cards-field
// @Summary      Bulk update card field
// @Description  Bulk-updates count or practiceCount on multiple cards owned by the authenticated user
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      models.UpdateCardsFieldRequest  true  "Bulk field update"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /update-cards-field [post]
func (h *Handler) UpdateCardsField(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.UpdateCardsFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.CardsService.UpdateField(c.Request.Context(), req, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"answer": "good"})
}

// DeleteCards handles POST /delete-cards
// @Summary      Delete cards
// @Description  Deletes cards by IDs owned by the authenticated user and unlinks them from parent memory nodes
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of card IDs"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /delete-cards [post]
func (h *Handler) DeleteCards(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.CardsService.DeleteMany(c.Request.Context(), req.IDs, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// CardsByQuery handles POST /cards-by-query
// @Summary      Filter cards by query
// @Description  Returns cards for a memory node (or priority/group), filtered by a simple query string
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request  body      models.CardsByQueryRequest  true  "Query request"
// @Success      200      {array}   models.CardFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /cards-by-query [post]
func (h *Handler) CardsByQuery(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.CardsByQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cards, err := h.App.CardsService.CardsByQuery(c.Request.Context(), req, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, []any{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// IncreaseCardCount handles PUT /increase-card-count/:id
// @Summary      Increase card count
// @Description  Increments the card count by 1
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /increase-card-count/{id} [put]
func (h *Handler) IncreaseCardCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	card, err := h.App.CardsService.IncrementCount(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// DecreaseCardCount handles PUT /decrease-card-count/:id
// @Summary      Decrease card count
// @Description  Decrements the card count by 1 (not below 0)
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /decrease-card-count/{id} [put]
func (h *Handler) DecreaseCardCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	card, err := h.App.CardsService.DecrementCount(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// IncreaseCardPracticeCount handles PUT /increase-card-practice-count/:id
// @Summary      Increase card practice count
// @Description  Increments the card practice count by 1
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /increase-card-practice-count/{id} [put]
func (h *Handler) IncreaseCardPracticeCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	card, err := h.App.CardsService.IncrementPracticeCount(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

// DecreaseCardPracticeCount handles PUT /decrease-card-practice-count/:id
// @Summary      Decrease card practice count
// @Description  Decrements the card practice count by 1 (not below 0)
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.CardFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /decrease-card-practice-count/{id} [put]
func (h *Handler) DecreaseCardPracticeCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	card, err := h.App.CardsService.DecrementPracticeCount(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}
