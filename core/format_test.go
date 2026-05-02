package core

import (
	"testing"

	"github.com/siherrmann/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJSON(t *testing.T) {
	v := validator.NewValidator()

	t.Run("Valid JSON", func(t *testing.T) {
		jsonStr := `{"summary": "This is a summary of length more than 10", "topics": ["test"]}`
		err := ValidateJSON(jsonStr, v)
		require.NoError(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		jsonStr := `{"summary": "incomplete"`
		err := ValidateJSON(jsonStr, v)
		assert.Error(t, err)
	})

	t.Run("Validation Error Summary", func(t *testing.T) {
		jsonStr := `{"summary": "short", "topics": ["test"]}`
		err := ValidateJSON(jsonStr, v)
		assert.Error(t, err)
	})

	t.Run("Validation Error Topics", func(t *testing.T) {
		jsonStr := `{"summary": "This is a summary of length more than 10", "topics": []}`
		err := ValidateJSON(jsonStr, v)
		assert.Error(t, err)
	})
}
