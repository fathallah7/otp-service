package queue

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeOTPDelivery = "otp:deliver"

type OTPDeliveryPayload struct {
	RequestID string `json:"request_id"`
	Recipient string `json:"recipient"`
}

func NewOTPDeliveryTask(
	requestID, recipient string,
) (*asynq.Task, error) {

	payload, err := json.Marshal(OTPDeliveryPayload{
		RequestID: requestID,
		Recipient: recipient,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal OTP task: %w", err)
	}

	return asynq.NewTask(TypeOTPDelivery, payload), nil
}
