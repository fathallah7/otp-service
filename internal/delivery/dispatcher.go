package delivery

import (
	"context"
	"fmt"

	"otp-service/internal/otp"
)

type Dispatcher struct {
	email    *EmailSender
	whatsapp *WhatsAppSender
}

func NewDispatcher(
	email *EmailSender,
	whatsapp *WhatsAppSender,
) *Dispatcher {
	return &Dispatcher{
		email:    email,
		whatsapp: whatsapp,
	}
}

func (d *Dispatcher) SendOTP(
	ctx context.Context,
	recipient otp.Recipient,
	code string,
) error {
	switch recipient.Channel {
	case otp.ChannelEmail:
		return d.email.SendOTP(ctx, recipient, code)

	case otp.ChannelWhatsApp:
		return d.whatsapp.SendOTP(ctx, recipient, code)

	default:
		return fmt.Errorf(
			"unsupported delivery channel: %s",
			recipient.Channel,
		)
	}
}
