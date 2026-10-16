package otp

import "testing"

func TestResolveRecipient(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantChannel Channel
		wantError   bool
	}{
		{
			name:        "Valid email",
			input:       "abdo@gmail.com",
			wantChannel: ChannelEmail,
		},
		{
			name:        "Valid phone",
			input:       "+14155552671",
			wantChannel: ChannelWhatsApp,
		},
		{
			name:      "Invalid email",
			input:     "abdo@@gmail.com",
			wantError: true,
		},
		{
			name:      "Invalid phone",
			input:     "12345",
			wantError: true,
		},
		{
			name:      "Empty recipient",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveRecipient(tt.input)

			if (err != nil) != tt.wantError {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil && result.Channel != tt.wantChannel {
				t.Errorf(
					"expected %s, got %s",
					tt.wantChannel,
					result.Channel,
				)
			}
		})
	}
}
