package handlers

import (
	"log"
	"net/http"
	"net/mail"
	"strings"

	"arturgudiev/memoryguard/models"
	"arturgudiev/memoryguard/services"

	"github.com/gin-gonic/gin"
)

// SendVerificationCode handles POST /admin/send-code
// @Summary      Send email verification code (admin)
// @Description  Admin-only. Generates a 6-digit OTP, returns it in the response, and sends it to the given email via SMTP.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      models.SendCodeRequest  true  "Recipient email"
// @Success      200      {object}  models.SendCodeResponse
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     Login[api]
// @Router       /admin/send-code [post]
func (h *Handler) SendVerificationCode(c *gin.Context) {
	admin, ok := h.isAdmin(c)
	if !ok {
		return
	}
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can send verification codes"})
		return
	}

	var req models.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	code, err := services.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate code"})
		return
	}

	if h.App.EmailService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email service is not configured"})
		return
	}

	if err := h.App.EmailService.SendVerificationCode(email, code); err != nil {
		log.Printf("SendVerificationCode: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
		return
	}

	c.JSON(http.StatusOK, models.SendCodeResponse{Code: code})
}
