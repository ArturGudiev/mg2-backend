package handlers

import (
	"arturgudiev/memoryguard/auth"
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func toUserResponse(user *ent.User) models.UserResponse {
	return models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
}

// GetUser handles GET /users/:id
// @Summary      Get a user by their ID
// @Description  Returns a user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "User ID"
// @Success      200      {object}  models.UserResponse
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}
	user, err := h.App.UsersRepository.GetUser(c.Request.Context(), idInt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, toUserResponse(user))
}

// GetMe handles GET /users/me
// @Summary      Get current user
// @Description  Returns the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  models.UserResponse
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	user, err := h.App.UsersRepository.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, toUserResponse(user))
}

// AddUser handles POST /users
// @Summary      Add a user
// @Description  Adds a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.NewUserRequest  true  "User to add"
// @Success      200      {object}  models.UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     []
// @Router       /users [post]
func (h *Handler) AddUser(c *gin.Context) {
	var req models.NewUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newUser, err := h.App.UsersRepository.AddUser(ctx, req.Name, req.Email, req.Password)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, toUserResponse(newUser))
}

// LoginUser handles POST /users/login
// @Summary      Logs in a user
// @Description  Logs in a user and returns tokens. Sets auth cookies.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.LoginUserRequest  true  "User to login"
// @Success      200      {object}  models.LoginUserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     []
// @Router       /users/login [post]
func (h *Handler) LoginUser(c *gin.Context) {
	var req models.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	foundUser, accessToken, refreshToken, err := h.authenticateAndIssueTokens(c, req.Email, req.Password)
	if err != nil {
		return
	}

	c.JSON(200, models.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(foundUser),
	})
}

// IssueToken handles POST /users/token (OAuth2 password grant for Swagger Authorize).
// @Summary      Issue OAuth2 access token
// @Description  OAuth2 password grant. Use email as username. Used by Swagger Authorize.
// @Tags         users
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        grant_type  formData  string  true  "Must be password"  Enums(password)
// @Param        username    formData  string  true  "User email"
// @Param        password    formData  string  true  "User password"
// @Success      200  {object}  models.OAuthTokenResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     []
// @Router       /users/token [post]
func (h *Handler) IssueToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	if grantType == "" {
		grantType = "password"
	}
	if grantType != "password" {
		c.JSON(400, gin.H{"error": "unsupported_grant_type"})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(400, gin.H{"error": "username and password are required"})
		return
	}

	_, accessToken, refreshToken, err := h.authenticateAndIssueTokens(c, username, password)
	if err != nil {
		return
	}

	c.JSON(200, models.OAuthTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    int(auth.AccessTokenTTL().Seconds()),
		RefreshToken: refreshToken,
	})
}

func (h *Handler) authenticateAndIssueTokens(c *gin.Context, email, password string) (*ent.User, string, string, error) {
	ctx := c.Request.Context()
	foundUser, err := h.App.UsersRepository.GetUserByCredentials(ctx, email, password)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(401, gin.H{"error": "Invalid email or password"})
			return nil, "", "", err
		}
		if ent.IsNotSingular(err) {
			c.JSON(500, gin.H{"error": "Multiple users found for this email"})
			return nil, "", "", err
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return nil, "", "", err
	}

	accessToken, refreshToken, refreshJTI, err := auth.GenerateTokenPair(foundUser.ID, foundUser.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return nil, "", "", err
	}

	if err := h.App.RefreshTokensRepository.CreateRefreshToken(ctx, refreshJTI, foundUser.ID, auth.RefreshTokenExpiry()); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return nil, "", "", err
	}

	auth.SetAuthCookies(c, accessToken, refreshToken)
	return foundUser, accessToken, refreshToken, nil
}

// RefreshUserToken handles POST /users/refresh
// @Summary      Refresh access token
// @Description  Exchanges a refresh token for a new access and refresh token pair
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshTokenRequest  false  "Refresh token"
// @Success      200      {object}  models.LoginUserResponse
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     []
// @Router       /users/refresh [post]
func (h *Handler) RefreshUserToken(c *gin.Context) {
	refreshToken := c.Query("refreshToken")
	if refreshToken == "" {
		if cookieToken, err := c.Cookie(auth.RefreshTokenCookieName); err == nil {
			refreshToken = cookieToken
		}
	}
	if refreshToken == "" {
		var req models.RefreshTokenRequest
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}
	if refreshToken == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "refresh token required"})
		return
	}

	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	ctx := c.Request.Context()
	storedToken, err := h.App.RefreshTokensRepository.GetActiveRefreshToken(ctx, claims.JTI)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "refresh token revoked or not found"})
		return
	}
	if storedToken.UserID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "refresh token mismatch"})
		return
	}
	if time.Now().UTC().After(storedToken.ExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "refresh token expired"})
		return
	}

	if err := h.App.RefreshTokensRepository.RevokeRefreshToken(ctx, claims.JTI); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	user, err := h.App.UsersRepository.GetUser(ctx, claims.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	accessToken, newRefreshToken, newRefreshJTI, err := auth.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := h.App.RefreshTokensRepository.CreateRefreshToken(ctx, newRefreshJTI, user.ID, auth.RefreshTokenExpiry()); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	auth.SetAuthCookies(c, accessToken, newRefreshToken)

	c.JSON(200, models.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         toUserResponse(user),
	})
}

// LogoutUser handles POST /users/logout
// @Summary      Logout user
// @Description  Revokes refresh token and clears auth cookies
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  map[string]string
// @Security     []
// @Router       /users/logout [post]
func (h *Handler) LogoutUser(c *gin.Context) {
	if refreshToken, err := c.Cookie(auth.RefreshTokenCookieName); err == nil && refreshToken != "" {
		if claims, err := auth.ValidateRefreshToken(refreshToken); err == nil {
			_ = h.App.RefreshTokensRepository.RevokeRefreshToken(c.Request.Context(), claims.JTI)
		}
	}

	auth.ClearAuthCookies(c)
	c.Status(http.StatusNoContent)
}
