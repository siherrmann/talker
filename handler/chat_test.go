package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
}
