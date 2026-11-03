package otp

type RequestOTPRequest struct {
	Recipient string `json:"recipient" binding:"required,max=254"`
}

type VerifyOTPRequest struct {
	Recipient string `json:"recipient" binding:"required,max=254"`
	Code      string `json:"code" binding:"required,len=6"`
}
