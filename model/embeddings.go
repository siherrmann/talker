package model

import (
	"encoding/json"
	"fmt"
)

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"-"` // Handled by custom unmarshaler
}

func (r *EmbeddingRequest) UnmarshalJSON(data []byte) error {
	type Alias EmbeddingRequest
	aux := &struct {
		Input json.RawMessage `json:"input"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Input) == 0 {
		return fmt.Errorf("input is required")
	}

	// Try unmarshaling as string
	var singleString string
	if err := json.Unmarshal(aux.Input, &singleString); err == nil {
		r.Input = []string{singleString}
		return nil
	}

	// Try unmarshaling as []string
	var arrayString []string
	if err := json.Unmarshal(aux.Input, &arrayString); err == nil {
		r.Input = arrayString
		return nil
	}

	return fmt.Errorf("input must be a string or an array of strings")
}

type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}
