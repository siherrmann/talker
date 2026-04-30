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

func TestEmbeddings(t *testing.T) {
	t.Run("ValidArrayInput", func(t *testing.T) {
		reqBody := []byte(`{
			"model": "text-embedding-ada-002",
			"input": ["First sentence", "Second sentence"]
		}`)

		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.EmbeddingResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "list", resp.Object)
		assert.Equal(t, "text-embedding-ada-002", resp.Model)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "embedding", resp.Data[0].Object)
		assert.Equal(t, 0, resp.Data[0].Index)
		assert.Len(t, resp.Data[0].Embedding, 3) // Mock returns 3 floats
		assert.Equal(t, float32(0.1), resp.Data[0].Embedding[0])
	})

	t.Run("ValidStringInput", func(t *testing.T) {
		reqBody := []byte(`{
			"model": "text-embedding-ada-002",
			"input": "Single sentence input"
		}`)

		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.EmbeddingResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Data, 1)
		assert.Len(t, resp.Data[0].Embedding, 3)
	})

	t.Run("EmptyInput", func(t *testing.T) {
		reqBody := []byte(`{
			"model": "text-embedding-ada-002",
			"input": []
		}`)

		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Input cannot be empty")
	})
}
