package waland

import "time"

const DefaultBaseURL = "https://api.waland.dev"

type SmsLogStatus string

const (
	SmsLogStatusPending SmsLogStatus = "pending"
	SmsLogStatusSent    SmsLogStatus = "sent"
	SmsLogStatusFailed  SmsLogStatus = "failed"
)

type ClientOptions struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient HTTPDoer
}

type SendMessageParams struct {
	ChatID        string
	Text          string
	MediaURL      string
	MediaFilename string
}

type SendMessageResult struct {
	ID             string       `json:"id"`
	SessionID      string       `json:"sessionId"`
	OrganizationID string       `json:"organizationId"`
	ChatID         string       `json:"chatId"`
	Text           *string      `json:"text"`
	MediaURL       *string      `json:"mediaUrl"`
	Status         SmsLogStatus `json:"status"`
	MessageID      *string      `json:"messageId"`
	Error          *string      `json:"error"`
	CreatedAt      string       `json:"createdAt"`
}
