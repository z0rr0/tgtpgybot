package ygpt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerationChat(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("failed content type header: %q", ct)
		}

		if ct := r.Header.Get("Accept"); ct != "application/json" {
			t.Errorf("failed content type header: %q", ct)
		}

		if auth := r.Header.Get("Authorization"); auth != "Api-Key test-key" {
			t.Errorf("failed authorization header: %q", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		response := `{"result":{"message":{"role":"Ассистент","text":"Меня зовут Алиса"},"num_tokens":"20"}}`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	client := s.Client()
	req := &ChatRequest{APIKey: "test-key", URL: s.URL, Text: "Кто ты?"}

	resp, err := GenerationChat(context.Background(), client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ChatResponse{
		Result: ChatResult{
			Message:      Message{Role: RoleAssistantRu, Text: "Меня зовут Алиса"},
			NumTokens:    "20",
			NumTokensInt: 20,
		},
	}

	if r := *resp; r != expected {
		t.Errorf("expected: %v, got: %v", expected, r)
	}
}

func TestGenerationChatFailedJSON(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		response := `{"result":{"message":{"role`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	client := s.Client()
	req := &ChatRequest{APIKey: "test-key", URL: s.URL, Text: "test"}

	_, err := GenerationChat(context.Background(), client, req)

	if !errors.Is(err, ErrChatGeneration) {
		t.Fatalf("expected error: %v, got: %v", ErrChatGeneration, err)
	}

	expectedPrefix := "failed to generate chat"
	if e := err.Error(); !strings.HasPrefix(e, expectedPrefix) {
		t.Fatalf("expected %q, got %q", expectedPrefix, e)
	}
}

func TestGenerationChatFailedStatus(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test", http.StatusBadGateway)
	}))
	defer s.Close()

	client := s.Client()
	req := &ChatRequest{APIKey: "test-key", URL: s.URL, Text: "test"}

	_, err := GenerationChat(context.Background(), client, req)

	if !errors.Is(err, ErrChatGeneration) {
		t.Fatalf("expected error: %v, got: %v", ErrChatGeneration, err)
	}

	expectedPrefix := "failed to generate chat\nunexpected status code"
	if e := err.Error(); !strings.HasPrefix(e, expectedPrefix) {
		t.Fatalf("expected %q, got %q", expectedPrefix, e)
	}
}

func TestGenerationChatMarshal(t *testing.T) {
	testCases := []struct {
		name      string
		req       ChatRequest
		expected  []string
		err       error
		errSubStr string
	}{
		{
			name:      "empty",
			req:       ChatRequest{},
			err:       ErrRequiredParam,
			errSubStr: "APIKey is empty",
		},
		{
			name:      "noAPIKey",
			req:       ChatRequest{URL: ChatURL, Text: "test"},
			err:       ErrRequiredParam,
			errSubStr: "APIKey is empty",
		},
		{
			name:      "noURL",
			req:       ChatRequest{APIKey: "test-key", Text: "test"},
			err:       ErrRequiredParam,
			errSubStr: "URL is empty",
		},
		{
			name:      "noText",
			req:       ChatRequest{APIKey: "test-key", URL: ChatURL},
			err:       ErrRequiredParam,
			errSubStr: "text is empty",
		},
		{
			name: "valid",
			req:  ChatRequest{APIKey: "test-key", URL: ChatURL, Text: "test"},
			expected: []string{
				`"model":"general"`,
				`"generationOptions":{`,
				`"partialResults":false`,
				`"temperature":1`,
				`"maxTokens":2000`,
				`"messages":[`,
				`"role":"User"`,
				`"text":"test"`,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			reader, err := tc.req.marshal()
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error: %v, got: %v", tc.err, err)
				}

				if !strings.Contains(err.Error(), tc.errSubStr) {
					t.Fatalf("expected error: %v, got: %v", tc.errSubStr, err)
				}

				return
			}

			if tc.err != nil {
				t.Fatalf("expected error, but got nil")
			}

			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			marshaled := string(data)
			for _, expected := range tc.expected {
				if !strings.Contains(marshaled, expected) {
					t.Errorf("expected: %q, got: %q", expected, marshaled)
				}
			}
		})
	}
}
