package handlers

import (
	"arturgudiev/memoryguard/auth"
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/services"
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var errAuthFailed = errors.New("authentication failed")

func toUserResponse(user *ent.User) models.UserResponse {
	return models.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Login:    user.Login,
		Email:    user.Email,
		Role:     user.Role,
		Verified: user.Verified,
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

// DeleteUser handles DELETE /admin/users/:id
// @Summary      Delete a user (admin)
// @Description  Admin-only. Deletes the user and their related data (nodes, cards, tokens). Admins cannot delete themselves.
// @Tags         admin
// @Produce      json
// @Param        id   path  int  true  "User ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Login[api]
// @Router       /admin/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	admin, ok := h.isAdmin(c)
	if !ok {
		return
	}
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can remove users"})
		return
	}

	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}

	adminID, ok := currentUserID(c)
	if !ok {
		return
	}
	if id == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	if err := h.App.UsersRepository.DeleteUser(c.Request.Context(), id); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
// @Summary      Register a user
// @Description  Creates an unverified user (login is optional) and emails a verification code. Existing unverified emails can re-register to update details and resend the code.
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

	name := strings.TrimSpace(req.Name)
	login := strings.TrimSpace(req.Login)
	email := strings.TrimSpace(req.Email)
	if name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(400, gin.H{"error": "invalid email"})
		return
	}

	ctx := c.Request.Context()
	existing, err := h.App.UsersRepository.GetUserByEmail(ctx, email)
	if err != nil && !ent.IsNotFound(err) {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if err == nil {
		if existing.Verified {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		updated, err := h.App.UsersRepository.UpdateUnverifiedRegistration(ctx, existing.ID, name, login, req.Password)
		if err != nil {
			if ent.IsConstraintError(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "email or login already in use"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if !h.sendRegistrationCode(c, updated) {
			return
		}
		h.maybeGrantSampleCards(ctx, updated.ID, req.AddSampleCards)
		c.JSON(200, toUserResponse(updated))
		return
	}

	newUser, err := h.App.UsersRepository.AddUser(ctx, name, login, email, req.Password, false)
	if err != nil {
		if ent.IsConstraintError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email or login already in use"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !h.sendRegistrationCode(c, newUser) {
		return
	}
	h.maybeGrantSampleCards(ctx, newUser.ID, req.AddSampleCards)

	c.JSON(200, toUserResponse(newUser))
}

// VerifyUser handles POST /users/verify
// @Summary      Verify a user email
// @Description  Confirms the registration code sent to the user's email and logs them in.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      models.VerifyUserRequest  true  "Email and verification code"
// @Success      200      {object}  models.LoginUserResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     []
// @Router       /users/verify [post]
func (h *Handler) VerifyUser(c *gin.Context) {
	var req models.VerifyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)
	ctx := c.Request.Context()

	foundUser, err := h.App.UsersRepository.GetUserByLoginOrEmail(ctx, email)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid verification code"})
		return
	}
	if foundUser.Verified {
		c.JSON(400, gin.H{"error": "email already verified"})
		return
	}
	vc, err := h.App.UsersRepository.GetVerificationCode(ctx, foundUser.ID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid verification code"})
		return
	}
	if time.Now().UTC().After(vc.ExpiresAt) {
		c.JSON(400, gin.H{"error": "verification code expired"})
		return
	}
	if len(vc.Code) != len(code) || subtle.ConstantTimeCompare([]byte(vc.Code), []byte(code)) != 1 {
		c.JSON(400, gin.H{"error": "invalid verification code"})
		return
	}

	if err := h.App.UsersRepository.MarkVerified(ctx, foundUser.ID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	foundUser.Verified = true

	accessToken, refreshToken, err := h.issueTokensForUser(c, foundUser)
	if err != nil {
		return
	}

	c.JSON(200, models.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(foundUser),
	})
}

// ResendVerificationCode handles POST /users/resend-code
// @Summary      Resend email verification code
// @Description  Sends a new verification code if the email belongs to an unverified account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      models.ResendCodeRequest  true  "Recipient email"
// @Success      200      {object}  map[string]bool
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     []
// @Router       /users/resend-code [post]
func (h *Handler) ResendVerificationCode(c *gin.Context) {
	var req models.ResendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	email := strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(400, gin.H{"error": "invalid email"})
		return
	}

	ctx := c.Request.Context()
	foundUser, err := h.App.UsersRepository.GetUserByLoginOrEmail(ctx, email)
	if err != nil {
		c.JSON(200, gin.H{"ok": true})
		return
	}
	if foundUser.Verified {
		c.JSON(200, gin.H{"ok": true})
		return
	}
	if !h.sendRegistrationCode(c, foundUser) {
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) sendRegistrationCode(c *gin.Context, user *ent.User) bool {
	if h.App.EmailService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email service is not configured"})
		return false
	}
	code, err := services.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate code"})
		return false
	}
	if err := h.App.UsersRepository.SetVerificationCode(c.Request.Context(), user.ID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if err := h.App.EmailService.SendVerificationCode(user.Email, code); err != nil {
		log.Printf("sendRegistrationCode: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
		return false
	}
	return true
}

func sampleMemoryNodeID() int {
	raw := strings.TrimSpace(os.Getenv("SAMPLE_MEMORY_NODE_ID"))
	if raw == "" {
		return 5
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		log.Printf("invalid SAMPLE_MEMORY_NODE_ID %q; using 5", raw)
		return 5
	}
	return id
}

func (h *Handler) maybeGrantSampleCards(ctx context.Context, userID int, add bool) {
	if !add {
		return
	}
	nodeID := sampleMemoryNodeID()
	if err := h.App.CardsService.MoveSharedNodeToUser(ctx, nodeID, userID, true); err != nil {
		log.Printf("grant sample memory node %d to user %d: %v", nodeID, userID, err)
	}
}

// LoginUser handles POST /users/login
// @Summary      Logs in a user
// @Description  Logs in a user by login or email and returns tokens. Sets auth cookies.
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
	if req.Identifier() == "" {
		c.JSON(400, gin.H{"error": "login or email is required"})
		return
	}

	foundUser, accessToken, refreshToken, err := h.authenticateAndIssueTokens(c, req.Identifier(), req.Password)
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
// @Description  OAuth2 password grant. Use login or email as username. Used by Swagger Authorize.
// @Tags         users
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        grant_type  formData  string  true  "Must be password"  Enums(password)
// @Param        username    formData  string  true  "User login or email"
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

func (h *Handler) authenticateAndIssueTokens(c *gin.Context, loginOrEmail, password string) (*ent.User, string, string, error) {
	ctx := c.Request.Context()
	foundUser, err := h.App.UsersRepository.GetUserByCredentials(ctx, loginOrEmail, password)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(401, gin.H{"error": "Invalid login or password"})
			return nil, "", "", err
		}
		if ent.IsNotSingular(err) {
			c.JSON(500, gin.H{"error": "Multiple users found for this login"})
			return nil, "", "", err
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return nil, "", "", err
	}
	if !foundUser.Verified {
		c.JSON(401, gin.H{"error": "email not verified", "email": foundUser.Email})
		return nil, "", "", errAuthFailed
	}

	accessToken, refreshToken, err := h.issueTokensForUser(c, foundUser)
	if err != nil {
		return nil, "", "", err
	}
	return foundUser, accessToken, refreshToken, nil
}

func (h *Handler) issueTokensForUser(c *gin.Context, foundUser *ent.User) (string, string, error) {
	accessToken, refreshToken, refreshJTI, err := auth.GenerateTokenPair(foundUser.ID, foundUser.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return "", "", err
	}

	if err := h.App.RefreshTokensRepository.CreateRefreshToken(c.Request.Context(), refreshJTI, foundUser.ID, auth.RefreshTokenExpiry()); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return "", "", err
	}

	auth.SetAuthCookies(c, accessToken, refreshToken)
	return accessToken, refreshToken, nil
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
