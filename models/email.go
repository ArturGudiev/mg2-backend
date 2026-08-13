package models

// SendCodeRequest is the body for POST /admin/send-code.
type SendCodeRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
}

// SendCodeResponse returns the generated OTP to the admin client.
type SendCodeResponse struct {
	Code string `json:"code" example:"482913"`
}
