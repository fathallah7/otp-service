package delivery

import (
	"context"
	"fmt"
	"strconv"

	"otp-service/internal/otp"

	"github.com/wneessen/go-mail"
)

type EmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailSender(host, port, username, password, from string) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailSender) SendOTP(
	ctx context.Context,
	recipient otp.Recipient,
	code string,
) error {
	// Only handle email recipients
	if recipient.Channel != otp.ChannelEmail {
		return fmt.Errorf("unsupported channel: %s", recipient.Channel)
	}

	// Create email message
	msg := mail.NewMsg()

	if err := msg.From(s.from); err != nil {
		return err
	}

	if err := msg.To(recipient.Value); err != nil {
		return err
	}

	msg.Subject("Your OTP Verification Code")

	msg.SetBodyString(
		mail.TypeTextPlain,
		fmt.Sprintf("Your verification code is: %s\nExpires in 5 minutes.", code),
	)

	// Convert SMTP port to integer
	port, err := strconv.Atoi(s.port)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	// Create SMTP client
	client, err := mail.NewClient(
		s.host,
		mail.WithPort(port),
		mail.WithTLSPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
	)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}

	// Send the email
	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send OTP email: %w", err)
	}

	return nil
}
