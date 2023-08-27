package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const tmpConfig = "/tmp/tgtpgybot_config_test.json"

func TestNew(t *testing.T) {
	_, err := New("/tmp/tgtpgbot_bad.json")
	if err == nil {
		t.Fatal("expected error")
	}

	cfg, err := New(tmpConfig)
	if err != nil {
		t.Fatal(err)
	}

	if cfg == nil {
		t.Fatal("config is nil")
	}

	expected := "token=****, timeout=1m0s, debug_level=info, chat.api_key=****, chat.proxy=empty"

	logValue := cfg.LogValue()
	if s := logValue.String(); s != expected {
		t.Errorf("log value is not equal: %q", s)
	}
}

func TestLoggerInit(t *testing.T) {
	cfg, err := New(tmpConfig)
	if err != nil {
		t.Fatal(err)
	}

	cfg.DebugLevel = "bad"
	if err = cfg.initLogger(); err == nil {
		t.Error("expected error")
	}

	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		cfg.DebugLevel = level

		if err = cfg.initLogger(); err != nil {
			t.Error(err)
		}
	}
}

func TestChatInit(t *testing.T) {
	cfg, err := New(tmpConfig)
	if err != nil {
		t.Fatal(err)
	}

	cfg.Chat.Proxy = "\n\t\r"
	cfg.Chat.Client = nil

	if err = cfg.Chat.init(); err == nil {
		t.Errorf("expected error: %#v", cfg.Chat)
	}

	cfg.Chat.Proxy = "https://127.0.0.1/proxy"
	cfg.Chat.Client = nil

	if err = cfg.Chat.init(); err != nil {
		t.Error(err)
	}

	cfg.Chat.Client = nil
	cfg.Chat.APIKey = ""

	if err = cfg.Chat.init(); err == nil {
		t.Errorf("expected error: %#v", cfg.Chat)
	}
}

func TestChatGeneration(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"result":{"message":{"role":"Ассистент","text":"Меня зовут Алиса"},"num_tokens":"20"}}`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	chat := &Chat{APIKey: "test-key", URL: s.URL, Client: s.Client()}
	expected := "Меня зовут Алиса"
	ctx := context.Background()

	value, err := chat.Generation(ctx, "Кто ты?", 1)
	if err != nil {
		t.Fatalf("failed to get completion: %v", err)
	}

	if value != expected {
		t.Errorf("completion value is not equal: %q", value)
	}
}
