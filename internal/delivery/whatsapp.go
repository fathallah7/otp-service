package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"otp-service/internal/otp"
	"strings"
	"time"
)

type WhatsAppSender struct {
	apiVersion       string
	phoneNumberID    string
	accessToken      string
	templateName     string
	templateLanguage string
	httpClient       *http.Client
}

func NewWhatsAppSender(
	apiVersion, phoneNumberID, accessToken,
	templateName, templateLanguage string,
) *WhatsAppSender {
	return &WhatsAppSender{
		apiVersion:       apiVersion,
		phoneNumberID:    phoneNumberID,
		accessToken:      accessToken,
		templateName:     templateName,
		templateLanguage: templateLanguage,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type WhatsAppMessage struct {
	MessagingProduct string           `json:"messaging_product"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Template         WhatsAppTemplate `json:"template"`
}

type WhatsAppTemplate struct {
	Name       string              `json:"name"`
	Language   WhatsAppLanguage    `json:"language"`
	Components []WhatsAppComponent `json:"components"`
}

type WhatsAppLanguage struct {
	Code string `json:"code"`
}

type WhatsAppComponent struct {
	Type       string              `json:"type"`
	Parameters []WhatsAppParameter `json:"parameters"`
}

type WhatsAppParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type WhatsAppAPIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (s *WhatsAppSender) buildTestPayload(to, code string) ([]byte, error) {
	payload := WhatsAppMessage{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "template",
		Template: WhatsAppTemplate{
			Name: s.templateName,
			Language: WhatsAppLanguage{
				Code: s.templateLanguage,
			},
			Components: []WhatsAppComponent{
				{
					Type: "body",
					Parameters: []WhatsAppParameter{
						{Type: "text", Text: "Test User"},
						{Type: "text", Text: code},
						{Type: "text", Text: "Oct 18, 2026"},
					},
				},
			},
		},
	}

	return json.Marshal(payload)
}

func (s *WhatsAppSender) buildTestRequest(
	ctx context.Context,
	to, code string,
) (*http.Request, error) {

	// 1. Build JSON payload
	payload, err := s.buildTestPayload(to, code)
	if err != nil {
		return nil, fmt.Errorf("build WhatsApp payload: %w", err)
	}

	// 2. Build Meta API URL
	url := fmt.Sprintf(
		"https://graph.facebook.com/%s/%s/messages",
		s.apiVersion,
		s.phoneNumberID,
	)

	// 3. Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create WhatsApp request: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	return req, nil
}

func (s *WhatsAppSender) sendTestRequest(
	ctx context.Context,
	to, code string,
) error {
	// 1. Build HTTP request
	req, err := s.buildTestRequest(ctx, to, code)
	if err != nil {
		return err
	}

	// 2. Send request to Meta
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send WhatsApp request: %w", err)
	}

	// 3. Close response body
	defer resp.Body.Close()

	// 4. Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr WhatsAppAPIError

		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf(
				"WhatsApp API returned HTTP %d: %w",
				resp.StatusCode,
				err,
			)
		}

		return fmt.Errorf(
			"WhatsApp API returned HTTP %d (code %d): %s",
			resp.StatusCode,
			apiErr.Error.Code,
			apiErr.Error.Message,
		)
	}

	// 5. Successful HTTP response
	return nil
}

func (s *WhatsAppSender) SendOTP(
	ctx context.Context,
	recipient otp.Recipient,
	code string,
) error {
	if recipient.Channel != otp.ChannelWhatsApp {
		return fmt.Errorf("unsupported channel: %s", recipient.Channel)
	}

	to := strings.TrimPrefix(recipient.Value, "+")

	return s.sendTestRequest(ctx, to, code)
}
