package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/repositories"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) (int, bool) {
	v, ok := c.Get("userID")
	if !ok {
		c.JSON(403, gin.H{"error": "unauthorized"})
		return 0, false
	}
	id, ok := v.(int)
	if !ok || id <= 0 {
		c.JSON(403, gin.H{"error": "unauthorized"})
		return 0, false
	}
	return id, true
}

func (h *Handler) currentUser(c *gin.Context) (*ent.User, bool) {
	userID, ok := currentUserID(c)
	if !ok {
		return nil, false
	}
	user, err := h.App.UsersRepository.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	return user, true
}

func (h *Handler) isAdmin(c *gin.Context) (bool, bool) {
	user, ok := h.currentUser(c)
	if !ok {
		return false, false
	}
	return user.Role == schema.UserRoleAdmin, true
}

// requireAdminForShared blocks non-admins from setting/changing the shared flag.
func (h *Handler) requireAdminForShared(c *gin.Context, changingShared bool) bool {
	if !changingShared {
		return true
	}
	admin, ok := h.isAdmin(c)
	if !ok {
		return false
	}
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can manage shared entities"})
		return false
	}
	return true
}

// Root handles GET /
// @Summary      Root endpoint
// @Description  Returns a welcome message
// @Tags         general
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       / [get]
func (h *Handler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "MemoryGuard server"})
}

func parsePositiveID(c *gin.Context, param string) (int, bool) {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return id, true
}

// GetMemoryNodeByID handles GET /memory-node/:id
// @Summary      Get memory node by ID
// @Description  Returns a memory node by its ID if the authenticated user owns it or has an explicit shared grant
// @Tags         memory-nodes
// @Produce      json
// @Param        id   path    int  true  "Memory node ID"
// @Success      200  {object}  models.MemoryNodeFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-node/{id} [get]
func (h *Handler) GetMemoryNodeByID(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	node, err := h.App.MemoryNodesService.Get(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// GetMemoryNodesByIDs handles POST /get-memory-nodes
// @Summary      Get memory nodes by IDs
// @Description  Returns multiple memory nodes by their IDs for the authenticated user
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of memory node IDs"
// @Success      200      {array}   models.MemoryNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /get-memory-nodes [post]
func (h *Handler) GetMemoryNodesByIDs(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	nodes, err := h.App.MemoryNodesService.GetByIDs(c.Request.Context(), req.IDs, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// ListMemoryNodes handles GET /memory-nodes
// @Summary      List memory nodes
// @Description  Returns all memory nodes for the authenticated user, or a subset when ids query is provided as a JSON array
// @Tags         memory-nodes
// @Produce      json
// @Param        ids  query     string  false  "JSON array of IDs, e.g. [1,2,3]"
// @Success      200  {array}   models.MemoryNodeFull
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-nodes [get]
func (h *Handler) ListMemoryNodes(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	if idsParam := c.Query("ids"); idsParam != "" {
		var ids []int
		if err := jsonNumberArray(idsParam, &ids); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ids query"})
			return
		}
		nodes, err := h.App.MemoryNodesService.GetByIDs(c.Request.Context(), ids, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nodes)
		return
	}

	nodes, err := h.App.MemoryNodesService.GetAll(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// ListRootMemoryNodes handles GET /memory-nodes/roots
// @Summary      List root memory nodes
// @Description  Returns memory nodes with no parents for the authenticated user (owned or explicitly granted shared roots)
// @Tags         memory-nodes
// @Produce      json
// @Success      200  {array}   models.MemoryNodeFull
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-nodes/roots [get]
func (h *Handler) ListRootMemoryNodes(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	nodes, err := h.App.MemoryNodesService.GetRoots(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// GetMemoryNodeByAlias handles GET /memory-node-by-alias/:alias and /node-by-alias/:alias
// @Summary      Get memory node by alias
// @Description  Finds a memory node by one of its aliases for the authenticated user
// @Tags         memory-nodes
// @Produce      json
// @Param        alias  path      string  true  "Alias"
// @Success      200    {object}  models.MemoryNodeFull
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-node-by-alias/{alias} [get]
func (h *Handler) GetMemoryNodeByAlias(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	alias := c.Param("alias")
	node, err := h.App.MemoryNodesService.GetByAlias(c.Request.Context(), alias, userID)
	if err != nil {
		if ent.IsNotFound(err) || errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// GetMemoryNodeParentsPath handles GET /memory-node/:id/parents-path
// @Summary      Get parents path for a memory node
// @Description  Returns the first-parent chain from root to the given node (inclusive)
// @Tags         memory-nodes
// @Produce      json
// @Param        id   path      int  true  "Memory node ID"
// @Success      200  {array}   models.MemoryNodePathItem
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-node/{id}/parents-path [get]
func (h *Handler) GetMemoryNodeParentsPath(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	path, err := h.App.MemoryNodesService.GetParentsPath(c.Request.Context(), id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, path)
}

// NewMemoryNode handles POST /new-memory-node
// @Summary      Create memory node
// @Description  Creates a new memory node owned by the authenticated user
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      models.NewMemoryNodeRequest  true  "Memory node creation request"
// @Success      200      {object}  models.MemoryNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /new-memory-node [post]
func (h *Handler) NewMemoryNode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.NewMemoryNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.MemoryNode.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory node payload"})
		return
	}
	if !h.requireAdminForShared(c, req.MemoryNode.Shared) {
		return
	}
	req.MemoryNode.UserID = userID
	node, err := h.App.MemoryNodesService.Create(c.Request.Context(), req.MemoryNode, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// UpdateMemoryNode handles PUT /update-memory-node
// @Summary      Update memory node
// @Description  Partially updates an existing memory node owned by the authenticated user
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      models.MemoryNodePartial  true  "Memory node update request"
// @Success      200      {object}  models.MemoryNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /update-memory-node [put]
func (h *Handler) UpdateMemoryNode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req models.MemoryNodePartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireAdminForShared(c, req.Shared != nil) {
		return
	}
	node, err := h.App.MemoryNodesService.Update(c.Request.Context(), req, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// DeleteMemoryNode handles DELETE /memory-node/:id
// @Summary      Delete memory node
// @Description  Deletes a memory node owned by the authenticated user. If the node is shared, also deletes owned child nodes, their cards, and all memory_node_users / card_users (and card_user_counts) links.
// @Tags         memory-nodes
// @Produce      json
// @Param        id   path      int  true  "Memory node ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-node/{id} [delete]
func (h *Handler) DeleteMemoryNode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	if err := h.App.MemoryNodesService.Delete(c.Request.Context(), id, userID); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// GrantMemoryNodeAccess handles POST /memory-node/:id/users
// @Summary      Grant shared memory node access
// @Description  Creates a memory_node_users row (and card_users for shared cards on the node) so the user can see the shared node. Does not make the node visible to everyone.
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        id       path      int                                true  "Memory node ID"
// @Param        request  body      models.GrantMemoryNodeAccessRequest true  "User to grant"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /memory-node/{id}/users [post]
func (h *Handler) GrantMemoryNodeAccess(c *gin.Context) {
	actorID, ok := currentUserID(c)
	if !ok {
		return
	}
	nodeID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req models.GrantMemoryNodeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	if err := h.App.CardsService.GrantSharedNodeAccess(c.Request.Context(), nodeID, req.UserID, actorID); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// MoveSharedNodeToUser handles POST /admin/memory-node/:id/move-to-user
// @Summary      Move shared memory node to user (admin)
// @Description  Admin-only. Grants a shared memory node and its shared cards to a user. When deep=true, also grants shared descendants recursively.
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        id       path      int                                   true  "Memory node ID"
// @Param        request  body      models.MoveSharedNodeToUserRequest    true  "Target user and deep flag"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /admin/memory-node/{id}/move-to-user [post]
func (h *Handler) MoveSharedNodeToUser(c *gin.Context) {
	admin, ok := h.isAdmin(c)
	if !ok {
		return
	}
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can move shared nodes to users"})
		return
	}
	nodeID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req models.MoveSharedNodeToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	if _, err := h.App.UsersRepository.GetUser(c.Request.Context(), req.UserID); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.CardsService.MoveSharedNodeToUser(c.Request.Context(), nodeID, req.UserID, req.Deep); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// RemoveSharedNodeFromUser handles POST /admin/memory-node/:id/remove-from-user
// @Summary      Remove shared memory node from user (admin)
// @Description  Admin-only. Revokes a user's access by deleting memory_node_users / card_users (and card_user_counts) links. Does not delete the node or cards. When deep=true, also revokes shared descendants recursively.
// @Tags         memory-nodes
// @Accept       json
// @Produce      json
// @Param        id       path      int                                      true  "Memory node ID"
// @Param        request  body      models.RemoveSharedNodeFromUserRequest   true  "Target user and deep flag"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /admin/memory-node/{id}/remove-from-user [post]
func (h *Handler) RemoveSharedNodeFromUser(c *gin.Context) {
	admin, ok := h.isAdmin(c)
	if !ok {
		return
	}
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can remove shared nodes from users"})
		return
	}
	nodeID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req models.RemoveSharedNodeFromUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	if _, err := h.App.UsersRepository.GetUser(c.Request.Context(), req.UserID); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.CardsService.RemoveSharedNodeFromUser(c.Request.Context(), nodeID, req.UserID, req.Deep); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory node not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
