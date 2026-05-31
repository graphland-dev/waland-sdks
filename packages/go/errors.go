package waland

import "strings"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type APIErrorBody struct {
	StatusCode int         `json:"statusCode"`
	Message    interface{} `json:"message"`
	Error      string      `json:"error,omitempty"`
}

type APIError struct {
	StatusCode int
	Message    string
	ErrorType  string
	Body       APIErrorBody
}

func (e *APIError) Error() string {
	return e.Message
}

func formatAPIMessage(message interface{}, fallback string) string {
	switch v := message.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fallback
		}
		return v
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			str, ok := item.(string)
			if !ok || strings.TrimSpace(str) == "" {
				continue
			}
			parts = append(parts, str)
		}
		if len(parts) == 0 {
			return fallback
		}
		return strings.Join(parts, "; ")
	default:
		return fallback
	}
}
