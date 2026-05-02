package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingRequest_UnmarshalJSON(t *testing.T) {
	t.Run("SingleStringInput", func(t *testing.T) {
		jsonData := []byte(`{"model": "test-model", "input": "hello world"}`)
		var req EmbeddingRequest
		err := json.Unmarshal(jsonData, &req)
		
		require.NoError(t, err)
		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, []string{"hello world"}, req.Input)
	})

	t.Run("ArrayStringInput", func(t *testing.T) {
		jsonData := []byte(`{"model": "test-model", "input": ["hello", "world"]}`)
		var req EmbeddingRequest
		err := json.Unmarshal(jsonData, &req)
		
		require.NoError(t, err)
		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, []string{"hello", "world"}, req.Input)
	})

	t.Run("InvalidTypeInput", func(t *testing.T) {
		jsonData := []byte(`{"model": "test-model", "input": 123}`)
		var req EmbeddingRequest
		err := json.Unmarshal(jsonData, &req)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input must be a string or an array of strings")
	})

	t.Run("MissingInput", func(t *testing.T) {
		jsonData := []byte(`{"model": "test-model"}`)
		var req EmbeddingRequest
		err := json.Unmarshal(jsonData, &req)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input is required")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		jsonData := []byte(`{"model": "test-model", "input": }`)
		var req EmbeddingRequest
		err := json.Unmarshal(jsonData, &req)
		
		assert.Error(t, err)
	})
}
