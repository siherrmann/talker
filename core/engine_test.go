package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errMock = errors.New("mock error")

func TestMockEngine_Generate(t *testing.T) {
	t.Run("DefaultResponse", func(t *testing.T) {
		m := &MockEngine{}
		resp, err := m.Generate(context.Background(), "hello", 10)
		require.NoError(t, err)
		assert.Equal(t, "This is a mock response to: hello", resp)
	})

	t.Run("CustomResponses", func(t *testing.T) {
		m := &MockEngine{
			Responses: []string{"resp1", "resp2"},
		}
		
		resp, err := m.Generate(context.Background(), "hello", 10)
		require.NoError(t, err)
		assert.Equal(t, "resp1", resp)
		
		resp2, err := m.Generate(context.Background(), "hello", 10)
		require.NoError(t, err)
		assert.Equal(t, "resp2", resp2)
		
		resp3, err := m.Generate(context.Background(), "hello", 10)
		require.NoError(t, err)
		assert.Equal(t, "resp2", resp3)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		m := &MockEngine{
			Err: errMock,
		}
		
		_, err := m.Generate(context.Background(), "hello", 10)
		assert.Error(t, err)
		assert.Equal(t, errMock, err)
	})
}

func TestMockEngine_GenerateStream(t *testing.T) {
	t.Run("DefaultResponse", func(t *testing.T) {
		m := &MockEngine{}
		tokenChan, errChan, err := m.GenerateStream(context.Background(), "hello", 10)
		require.NoError(t, err)
		
		var tokens []string
		done := false
		
		for !done {
			select {
			case token, ok := <-tokenChan:
				if !ok {
					done = true
				} else {
					tokens = append(tokens, token)
				}
			case err := <-errChan:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for stream")
			}
		}
		
		expectedTokens := []string{"This", " ", "is", " ", "a", " ", "mock", " ", "response", " ", "to:", " ", "hello"}
		assert.Equal(t, expectedTokens, tokens)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		m := &MockEngine{
			Err: errMock,
		}
		_, _, err := m.GenerateStream(context.Background(), "hello", 10)
		assert.Error(t, err)
		assert.Equal(t, errMock, err)
	})
}

func TestMockEngine_ExtractEmbeddings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &MockEngine{}
		res, err := m.ExtractEmbeddings(context.Background(), []string{"test1", "test2"})
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, []float32{0.1, 0.2, 0.3}, res[0])
		assert.Equal(t, []float32{0.1, 0.2, 0.3}, res[1])
	})

	t.Run("Error", func(t *testing.T) {
		m := &MockEngine{Err: errMock}
		_, err := m.ExtractEmbeddings(context.Background(), []string{"test1"})
		assert.Error(t, err)
	})
}

func TestMockEngine_Close(t *testing.T) {
	m := &MockEngine{}
	err := m.Close()
	require.NoError(t, err)
}

func TestNewHugotEngine_InvalidPath(t *testing.T) {
	// Attempt to create HugotEngine with an invalid path (permission denied typically) to force error
	// For testing purposes, we use a root directory we can't write to.
	_, err := NewHugotEngine("/root/forbidden_dir_test", "", "")
	assert.Error(t, err)
}
