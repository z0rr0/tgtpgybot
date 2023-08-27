package ygpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// ChatURL is a chat generation API URL.
const ChatURL = "https://llm.api.cloud.yandex.net/llm/v1alpha/chat"

var (
	// ErrRequiredParam is an error that occurs when a required parameter is missing.
	ErrRequiredParam = errors.New("required parameter is missing")

	// ErrChatGeneration is an error that occurs when a chat generation request fails.
	ErrChatGeneration = errors.New("failed to generate chat")
)

// GenerationOptions is a model configuration parameters.
type GenerationOptions struct {
	PartialResults bool    `json:"partialResults"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      int64   `json:"maxTokens"`
}

// Message is a chat message.
type Message struct {
	Role Role   `json:"role"`
	Text string `json:"text"`
}

// TextGenerationChat is a request to the chat generation API.
type TextGenerationChat struct {
	Model             Model             `json:"model"`
	GenerationOptions GenerationOptions `json:"generationOptions"`
	Messages          []Message         `json:"messages"`
	InstructionText   string            `json:"instructionText,omitempty"`
}

// ChatResult is a result of the chat generation API.
type ChatResult struct {
	Message      Message `json:"message"`
	NumTokens    string  `json:"num_tokens"` // int64 string
	NumTokensInt int64   `json:"-"`          // parsed NumTokens
}

// ChatResponse is a response from the chat generation API.
type ChatResponse struct {
	Result ChatResult `json:"result"`
}

// String implements the fmt.Stringer interface.
func (cr *ChatResponse) String() string {
	return cr.Result.Message.Text
}

// parseNumTokens parses NumTokens to NumTokensInt.
func (cr *ChatResponse) parseNumTokens() error {
	numTokens, err := strconv.ParseInt(cr.Result.NumTokens, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse numTokens: %w", err)
	}

	cr.Result.NumTokensInt = numTokens
	return nil
}

// ChatRequest is a request params structure for the chat generation API.
type ChatRequest struct {
	APIKey string
	URL    string
	Text   string
}

func (c *ChatRequest) validate() error {
	if c.APIKey == "" {
		return errors.Join(ErrRequiredParam, fmt.Errorf("APIKey is empty"))
	}

	if c.URL == "" {
		return errors.Join(ErrRequiredParam, fmt.Errorf("URL is empty"))
	}

	if c.Text == "" {
		return errors.Join(ErrRequiredParam, fmt.Errorf("text is empty"))
	}

	return nil
}

// marshal returns a reader with the request body.
func (c *ChatRequest) marshal() (io.Reader, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	// YandexGPT API is preview, so use only "general" model, Temperature=0 and MaxTokens=2000.
	chatData := &TextGenerationChat{
		Model:             ModelGeneral,
		GenerationOptions: GenerationOptions{MaxTokens: 2000},
		Messages:          []Message{{Role: RoleUser, Text: c.Text}},
	}

	data, err := json.Marshal(chatData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return bytes.NewReader(data), nil
}

// build returns a new http.Request.
func (c *ChatRequest) build(ctx context.Context) (*http.Request, error) {
	data, err := c.marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.APIKey)

	return req, nil
}

// GenerationChat returns a new chat generation response.
func GenerationChat(ctx context.Context, client *http.Client, req *ChatRequest) (*ChatResponse, error) {
	request, err := req.build(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.Join(ErrChatGeneration, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, statusError(resp.StatusCode, resp.Body)
	}

	return buildResponse(resp.Body)
}

func statusError(status int, body io.Reader) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return errors.Join(
			ErrChatGeneration,
			fmt.Errorf("unexpected status code=%d", status),
			err,
		)
	}

	return errors.Join(
		ErrChatGeneration,
		fmt.Errorf("unexpected status code=%v: %v", status, string(bodyBytes)),
	)
}

func buildResponse(reader io.Reader) (*ChatResponse, error) {
	response := &ChatResponse{}
	if err := json.NewDecoder(reader).Decode(response); err != nil {
		return nil, errors.Join(ErrChatGeneration, err)
	}

	if err := response.parseNumTokens(); err != nil {
		return nil, errors.Join(ErrChatGeneration, err)
	}

	return response, nil
}
