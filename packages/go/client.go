package waland

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var chatIDPattern = regexp.MustCompile(`^[^@\s]+@(s\.whatsapp\.net|g\.us)$`)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	apiKey    string
	sessionID string
	baseURL   string
	timeout   time.Duration
	http      HTTPDoer
}

func NewClient(apiKey, sessionID string, options *ClientOptions) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, &ValidationError{Message: "apiKey is required"}
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, &ValidationError{Message: "sessionId is required"}
	}

	baseURL := DefaultBaseURL
	timeout := 30 * time.Second
	var httpClient HTTPDoer = &http.Client{Timeout: timeout}

	if options != nil {
		if strings.TrimSpace(options.BaseURL) != "" {
			baseURL = strings.TrimSuffix(strings.TrimSpace(options.BaseURL), "/")
		}
		if options.Timeout > 0 {
			timeout = options.Timeout
		}
		if options.HTTPClient != nil {
			httpClient = options.HTTPClient
		}
	}

	if options == nil || options.HTTPClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		apiKey:    strings.TrimSpace(apiKey),
		sessionID: strings.TrimSpace(sessionID),
		baseURL:   baseURL,
		timeout:   timeout,
		http:      httpClient,
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, params SendMessageParams) (SendMessageResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := validateSendMessageParams(params); err != nil {
		return SendMessageResult{}, err
	}

	payload := map[string]string{
		"chatId": strings.TrimSpace(params.ChatID),
	}

	if text := strings.TrimSpace(params.Text); text != "" {
		payload["text"] = text
	}
	if mediaURL := strings.TrimSpace(params.MediaURL); mediaURL != "" {
		payload["mediaUrl"] = mediaURL
	}
	if mediaFilename := strings.TrimSpace(params.MediaFilename); mediaFilename != "" {
		payload["mediaFilename"] = mediaFilename
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return SendMessageResult{}, err
	}

	requestURL := fmt.Sprintf("%s/v1/sessions/%s/send", c.baseURL, url.PathEscape(c.sessionID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return SendMessageResult{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	response, err := c.http.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return SendMessageResult{}, &APIError{
				StatusCode: http.StatusRequestTimeout,
				Message:    fmt.Sprintf("Request timed out after %dms", c.timeout.Milliseconds()),
				ErrorType:  "Request Timeout",
				Body: APIErrorBody{
					StatusCode: http.StatusRequestTimeout,
					Message:    fmt.Sprintf("Request timed out after %dms", c.timeout.Milliseconds()),
					Error:      "Request Timeout",
				},
			}
		}
		return SendMessageResult{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return SendMessageResult{}, err
	}

	parsedBody := parseJSONBody(responseBody, response.StatusCode)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		normalized := normalizeErrorBody(response.StatusCode, parsedBody)
		return SendMessageResult{}, &APIError{
			StatusCode: normalized.StatusCode,
			Message:    formatAPIMessage(normalized.Message, statusMessage(response.StatusCode)),
			ErrorType:  normalized.Error,
			Body:       normalized,
		}
	}

	var result SendMessageResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return SendMessageResult{}, err
	}

	return result, nil
}

func validateSendMessageParams(params SendMessageParams) error {
	if strings.TrimSpace(params.ChatID) == "" {
		return &ValidationError{Message: "chatId is required"}
	}

	if !chatIDPattern.MatchString(strings.TrimSpace(params.ChatID)) {
		return &ValidationError{Message: "chatId must be a WhatsApp JID, e.g. 8801712345678@s.whatsapp.net or {groupId}@g.us"}
	}

	text := strings.TrimSpace(params.Text)
	mediaURL := strings.TrimSpace(params.MediaURL)
	if text == "" && mediaURL == "" {
		return &ValidationError{Message: "Either text or mediaUrl must be provided"}
	}

	if mediaURL != "" {
		parsed, err := url.Parse(mediaURL)
		if err != nil || parsed.Host == "" {
			return &ValidationError{Message: "mediaUrl must be a valid URL"}
		}
		if !strings.HasPrefix(parsed.Scheme, "http") {
			return &ValidationError{Message: "mediaUrl must include a protocol (http or https)"}
		}
	}

	if params.MediaFilename != "" && strings.TrimSpace(params.MediaFilename) == "" {
		return &ValidationError{Message: "mediaFilename cannot be empty"}
	}

	return nil
}

func parseJSONBody(body []byte, statusCode int) map[string]interface{} {
	if len(body) == 0 {
		return map[string]interface{}{}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return map[string]interface{}{
			"statusCode": statusCode,
			"message":    string(body),
			"error":      "Error",
		}
	}
	return payload
}

func normalizeErrorBody(status int, payload map[string]interface{}) APIErrorBody {
	if _, ok := payload["statusCode"]; ok {
		statusCode := status
		if rawStatus, ok := payload["statusCode"].(float64); ok {
			statusCode = int(rawStatus)
		}

		message, hasMessage := payload["message"]
		if !hasMessage {
			message = statusMessage(status)
		}

		errorType := ""
		if rawError, ok := payload["error"].(string); ok {
			errorType = rawError
		}

		return APIErrorBody{
			StatusCode: statusCode,
			Message:    message,
			Error:      errorType,
		}
	}

	message := interface{}(statusMessage(status))
	if rawMessage, ok := payload["message"]; ok {
		message = rawMessage
	}

	return APIErrorBody{
		StatusCode: status,
		Message:    message,
		Error:      "Error",
	}
}

func statusMessage(status int) string {
	return fmt.Sprintf("Request failed with status %d", status)
}

func isTimeoutError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if errors.Is(err, context.Canceled) {
		return false
	}

	return false
}
