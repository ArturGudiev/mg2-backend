package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

func cookieSameSite() http.SameSite {
	if cookieSecure() {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := cookieSecure()
	sameSite := cookieSameSite()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(AccessTokenTTL().Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int(RefreshTokenTTL().Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ClearAuthCookies(c *gin.Context) {
	secure := cookieSecure()
	sameSite := cookieSameSite()
	expired := time.Unix(0, 0)

	for _, name := range []string{AccessTokenCookieName, RefreshTokenCookieName} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  expired,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})
	}
}
