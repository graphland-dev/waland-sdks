package waland

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testAPIKey    = "waland_test_key"
	testSessionID = "session-abc123"
)

func TestNewClientRequiresCredentials(t *testing.T) {
	if _, err := NewClient("", testSessionID, nil); err == nil {
		t.Fatalf("expected error for empty api key")
	}

	if _, err := NewClient(testAPIKey, "", nil); err == nil {
		t.Fatalf("expected error for empty session ID")
	}
}

func TestSendTextMessage(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sessions/session-abc123/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"log-id","sessionId":"session-abc123","organizationId":"org-id","chatId":"8801712345678@s.whatsapp.net","text":"Hello","mediaUrl":null,"status":"sent","messageId":"wa-msg-id","error":null,"createdAt":"2026-05-24T10:00:00.000Z"}`))
	}))
	defer server.Close()

	client, err := NewClient(testAPIKey, testSessionID, &ClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	result, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID: "8801712345678@s.whatsapp.net",
		Text:   "Hello",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if gotAuth != "Bearer "+testAPIKey {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if gotBody["chatId"] != "8801712345678@s.whatsapp.net" || gotBody["text"] != "Hello" {
		t.Fatalf("unexpected request body: %#v", gotBody)
	}
	if result.Status != SmsLogStatusSent {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestSendMediaMessage(t *testing.T) {
	t.Parallel()

	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"log-id","sessionId":"session-abc123","organizationId":"org-id","chatId":"8801712345678@s.whatsapp.net","text":"Caption","mediaUrl":"https://example.com/photo.jpg","status":"sent","messageId":"wa-msg-id","error":null,"createdAt":"2026-05-24T10:00:00.000Z"}`))
	}))
	defer server.Close()

	client, _ := NewClient(testAPIKey, testSessionID, &ClientOptions{BaseURL: server.URL})
	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID:        "8801712345678@s.whatsapp.net",
		Text:          "Caption",
		MediaURL:      "https://example.com/photo.jpg",
		MediaFilename: "photo.jpg",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if gotBody["mediaUrl"] != "https://example.com/photo.jpg" {
		t.Fatalf("missing mediaUrl: %#v", gotBody)
	}
	if gotBody["mediaFilename"] != "photo.jpg" {
		t.Fatalf("missing mediaFilename: %#v", gotBody)
	}
}

func TestRejectsInvalidChatID(t *testing.T) {
	client, _ := NewClient(testAPIKey, testSessionID, nil)

	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID: "not-a-jid",
		Text:   "Hi",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestRejectsMissingTextAndMedia(t *testing.T) {
	client, _ := NewClient(testAPIKey, testSessionID, nil)

	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID: "8801712345678@s.whatsapp.net",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"statusCode":401,"message":"Invalid or missing org API key","error":"Unauthorized"}`))
	}))
	defer server.Close()

	client, _ := NewClient(testAPIKey, testSessionID, &ClientOptions{BaseURL: server.URL})
	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID: "8801712345678@s.whatsapp.net",
		Text:   "Hi",
	})
	if err == nil {
		t.Fatalf("expected API error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("unexpected status code: %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid or missing org API key" {
		t.Fatalf("unexpected message: %s", apiErr.Message)
	}
}

func TestCheckNumber(t *testing.T) {
	t.Parallel()

	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sessions/session-abc123/check-number" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"number":"8801712345678","chatId":"8801712345678@s.whatsapp.net","jid":"8801712345678@s.whatsapp.net","exists":true}`))
	}))
	defer server.Close()

	client, _ := NewClient(testAPIKey, testSessionID, &ClientOptions{BaseURL: server.URL})
	result, err := client.CheckNumber(context.Background(), CheckNumberParams{Number: "8801712345678"})
	if err != nil {
		t.Fatalf("CheckNumber failed: %v", err)
	}

	if gotBody["number"] != "8801712345678" {
		t.Fatalf("unexpected request body: %#v", gotBody)
	}
	if result.Exists == nil || !*result.Exists {
		t.Fatalf("expected exists=true, got %#v", result.Exists)
	}
}

func TestCheckNumberRequiresNumber(t *testing.T) {
	client, _ := NewClient(testAPIKey, testSessionID, nil)

	_, err := client.CheckNumber(context.Background(), CheckNumberParams{Number: " "})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}
