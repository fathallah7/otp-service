package otp

type RequestOTPRequest struct {
	Recipient string `json:"recipient" binding:"required,max=254"`
}
