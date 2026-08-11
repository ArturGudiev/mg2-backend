package handlers

import (
	"net/http"

	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/models"

	"github.com/gin-gonic/gin"
)

// GetCardItemByID handles GET /card-item/:id
// @Summary      Get card item by ID
// @Description  Returns a card item by its ID for the authenticated user
// @Tags         card-items
// @Produce      json
// @Param        id   path      int  true  "Card item ID"
// @Success      200  {object}  models.CardItemFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /card-item/{id} [get]
func (h *Handler) GetCardItemByID(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	item, err := h.App.CardItemsService.Get(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Card item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// GetCardItemsByIDs handles POST /get-card-items
// @Summary      Get card items by IDs
// @Description  Returns multiple card items by their IDs for the authenticated user
// @Tags         card-items
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of card item IDs"
// @Success      200      {array}   models.CardItemFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /get-card-items [post]
func (h *Handler) GetCardItemsByIDs(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.App.CardItemsService.GetByIDs(c.Request.Context(), req.IDs, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// ListCardItems handles GET /card-items
// @Summary      List card items
// @Description  Returns all card items for the authenticated user
// @Tags         card-items
// @Produce      json
// @Success      200  {array}   models.CardItemFull
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /card-items [get]
func (h *Handler) ListCardItems(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	items, err := h.App.CardItemsService.GetAll(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// NewCardItem handles POST /new-card-item
// @Summary      Create card item
// @Description  Creates a new card item owned by the authenticated user
// @Tags         card-items
// @Accept       json
// @Produce      json
// @Param        request  body      models.CardItemShort  true  "Card item"
// @Success      200      {object}  models.CardItemFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Security     Login[api]
// @Router       /new-card-item [post]
func (h *Handler) NewCardItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.CardItemShort
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireAdminForShared(c, req.Shared != nil && *req.Shared) {
		return
	}
	item, err := h.App.CardItemsService.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpdateCardItem handles PUT /update-card-item
// @Summary      Update card item
// @Description  Partially updates an existing card item owned by the authenticated user
// @Tags         card-items
// @Accept       json
// @Produce      json
// @Param        request  body      models.CardItemPartial  true  "Card item update"
// @Success      200      {object}  models.CardItemFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /update-card-item [put]
func (h *Handler) UpdateCardItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.CardItemPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireAdminForShared(c, req.Shared != nil) {
		return
	}
	item, err := h.App.CardItemsService.Update(c.Request.Context(), req, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Card item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteCardItem handles DELETE /card-item/:id
// @Summary      Delete card item
// @Description  Deletes a card item by ID owned by the authenticated user
// @Tags         card-items
// @Produce      json
// @Param        id   path      int  true  "Card item ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /card-item/{id} [delete]
func (h *Handler) DeleteCardItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	if err := h.App.CardItemsService.Delete(c.Request.Context(), id, userID); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Card item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
