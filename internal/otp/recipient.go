package otp

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/nyaruka/phonenumbers/v2"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelWhatsApp Channel = "whatsapp"
)

var ErrInvalidRecipient = errors.New("invalid recipient")

type Recipient struct {
	Value   string
	Channel Channel
}

func ResolveRecipient(input string) (Recipient, error) {
	value := strings.TrimSpace(input)

	if value == "" || len(value) > 254 {
		return Recipient{}, ErrInvalidRecipient
	}

	// Email detection
	if strings.Contains(value, "@") {
		address, err := mail.ParseAddress(value)

		if err != nil || address.Address != value || address.Name != "" {
			return Recipient{}, ErrInvalidRecipient
		}

		at := strings.LastIndexByte(value, '@')
		normalized := value[:at+1] + strings.ToLower(value[at+1:])

		return Recipient{
			Value:   normalized,
			Channel: ChannelEmail,
		}, nil
	}

	// Phone
	if !strings.HasPrefix(value, "+") {
		return Recipient{}, ErrInvalidRecipient
	}

	number, err := phonenumbers.Parse(value, "ZZ")

	if err != nil || !phonenumbers.IsValidNumber(number) {
		return Recipient{}, ErrInvalidRecipient
	}

	return Recipient{
		Value:   phonenumbers.Format(number, phonenumbers.E164),
		Channel: ChannelWhatsApp,
	}, nil

}
