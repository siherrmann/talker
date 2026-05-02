package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/core"

	"github.com/siherrmann/talker/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatCompletions(t *testing.T) {
	t.Run("ValidRequest", func(t *testing.T) {
		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
			MaxTokens: 50,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.ChatCompletionResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "chat.completion", resp.Object)
		assert.Equal(t, "test-model", resp.Model)
		assert.Len(t, resp.Choices, 1)
		assert.Contains(t, resp.Choices[0].Message.Content, "This is a mock response to:")
	})

	t.Run("EmptyMessages", func(t *testing.T) {
		reqBody := model.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []model.ChatMessage{},
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Messages array cannot be empty")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid request payload")
	})

	t.Run("JSONValidationRetrySuccess", func(t *testing.T) {
		// Mock engine to return invalid JSON first, then invalid validation, then valid JSON
		mockEngine.Responses = []string{
			"invalid json",
			`{"summary": "short", "topics": []}`, // Valid JSON but invalid validator (min10, min1)
			`{"summary": "This is a long enough summary.", "topics": ["test"]}`, // Valid
		}
		defer func() { mockEngine.Responses = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
			ResponseFormat: &model.ResponseFormat{Type: "json_object"},
			MaxTokens:      50,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.ChatCompletionResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, `{"summary": "This is a long enough summary.", "topics": ["test"]}`, resp.Choices[0].Message.Content)
	})

	t.Run("Streaming", func(t *testing.T) {
		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Stream this"},
			},
			Stream:    true,
			MaxTokens: 50,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

		// Ensure we have some streaming output
		body := rec.Body.String()
		assert.Contains(t, body, "data: {")
		assert.Contains(t, body, "chat.completion.chunk")
		assert.Contains(t, body, "data: [DONE]")
	})

	t.Run("NonStreamingError_Generate", func(t *testing.T) {
		mockEngine.Err = errors.New("mock engine error")
		defer func() { mockEngine.Err = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to generate completion")
	})

	t.Run("NonStreamingError_JSON_Generate", func(t *testing.T) {
		mockEngine.Err = errors.New("mock engine error json")
		defer func() { mockEngine.Err = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{{Role: "user", Content: "Hello!"}},
			ResponseFormat: &model.ResponseFormat{Type: "json_object"},
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to generate completion")
	})

	t.Run("StreamingError_GenerateStream", func(t *testing.T) {
		mockEngine.Err = errors.New("mock stream error")
		defer func() { mockEngine.Err = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Stream error"},
			},
			Stream: true,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to generate completion stream")
	})

	t.Run("StreamingError_StreamChan", func(t *testing.T) {
		mockEngine.StreamErr = errors.New("channel error")
		defer func() { mockEngine.StreamErr = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{{Role: "user", Content: "Stream error"}},
			Stream: true,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// The stream has already started, so the status code is 200, 
		// The stream error bubbles up, so the status code becomes 500
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("JSONValidationRetryFailure", func(t *testing.T) {
		// Mock engine to return invalid JSON enough times to fail
		mockEngine.Responses = []string{
			"invalid json 1",
			"invalid json 2",
			"invalid json 3",
		}
		defer func() { mockEngine.Responses = nil }()

		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
			ResponseFormat: &model.ResponseFormat{Type: "json_object"},
			MaxTokens:      50,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to generate valid JSON after multiple attempts")
	})

	t.Run("StreamingContextCanceled", func(t *testing.T) {
		reqBody := model.ChatCompletionRequest{
			Model: "test-model",
			Messages: []model.ChatMessage{
				{Role: "user", Content: "Stream this"},
			},
			Stream: true,
		}

		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// Cancel the request context to trigger the Done() path in the stream loop
		ctx, cancel := context.WithCancel(req.Context())
		req = req.WithContext(ctx)
		cancel()

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code) // Headers might be set before context cancels
	})
}

// Custom recorder to simulate a Write error
type errorRecorder struct {
	*httptest.ResponseRecorder
}

func (e *errorRecorder) Write(buf []byte) (int, error) {
	return 0, errors.New("mock write error")
}

func TestChatCompletions_StreamWriteError(t *testing.T) {
	reqBody := model.ChatCompletionRequest{
		Model: "test-model",
		Messages: []model.ChatMessage{{Role: "user", Content: "Stream this"}},
		Stream: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	
	rec := &errorRecorder{httptest.NewRecorder()}
	
	e := echo.New()
	handler := NewChatHandler(&core.MockEngine{})
	e.POST("/v1/chat/completions", handler.ChatCompletions)
	e.ServeHTTP(rec, req)
}
