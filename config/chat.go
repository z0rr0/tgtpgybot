package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/z0rr0/tgtpgybot/ygpt"
)

// Chat is a chat generation API configuration.
type Chat struct {
	APIKey string       `json:"api_key"`
	Proxy  string       `json:"proxy"`
	URL    string       `json:"-"`
	Client *http.Client `json:"-"`
}

// init creates a new HTTP client and sets the chat generation API URL.
func (chat *Chat) init() error {
	if chat.Client != nil {
		return nil
	}

	if chat.APIKey == "" {
		return fmt.Errorf("empty API key")
	}

	if chat.Proxy != "" {
		proxyURL, err := url.Parse(chat.Proxy)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		chat.Client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	} else {
		chat.Client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	}

	chat.URL = ygpt.ChatURL
	return nil
}

// Generation generates a new GPT text response.
func (chat *Chat) Generation(ctx context.Context, text string) (string, error) {
	request := &ygpt.ChatRequest{
		APIKey: chat.APIKey,
		URL:    chat.URL,
		Text:   text,
	}

	resp, err := ygpt.GenerationChat(ctx, chat.Client, request)
	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	return resp.String(), nil
}
