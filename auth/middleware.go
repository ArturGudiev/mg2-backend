package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func isPublicRoute(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}
	if strings.HasPrefix(path, "/swagger") {
		return true
	}

	switch {
	case method == http.MethodPost && path == "/users/login":
		return true
	case method == http.MethodPost && path == "/users/token":
		return true
	case method == http.MethodPost && path == "/users":
		return true
	case method == http.MethodPost && path == "/users/verify":
		return true
	case method == http.MethodPost && path == "/users/resend-code":
		return true
	case method == http.MethodPost && path == "/users/refresh":
		return true
	case method == http.MethodPost && path == "/users/logout":
		return true
	case method == http.MethodGet && path == "/":
		return true
	}

	return false
}

func extractAccessToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	accessToken, err := c.Cookie(AccessTokenCookieName)
	if err != nil {
		return ""
	}
	return accessToken
}

// AuthMiddleware requires a valid Bearer token or access_token cookie on protected routes.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicRoute(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		accessToken := extractAccessToken(c)
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access token required"})
			return
		}

		claims, err := ValidateAccessToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid or expired access token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
