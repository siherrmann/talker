package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/core"
	"github.com/siherrmann/talker/metrics"
	"github.com/siherrmann/talker/model"
)

type EmbeddingsHandler struct {
	Engine core.Engine
}

func NewEmbeddingsHandler(engine core.Engine) *EmbeddingsHandler {
	return &EmbeddingsHandler{
		Engine: engine,
	}
}

func (h *EmbeddingsHandler) Embeddings(c *echo.Context) error {
	var req model.EmbeddingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload: " + err.Error()})
	}

	if len(req.Input) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Input cannot be empty"})
	}

	embeddings, err := h.Engine.ExtractEmbeddings(c.Request().Context(), req.Input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to extract embeddings: " + err.Error()})
	}

	data := make([]model.Embedding, len(embeddings))
	for i, emb := range embeddings {
		data[i] = model.Embedding{
			Object:    "embedding",
			Embedding: emb,
			Index:     i,
		}
	}

	// Calculate exact usage using tokenizers
	promptTokens := 0
	for _, text := range req.Input {
		count, _ := h.Engine.CountTokens(text, true)
		promptTokens += count
	}

	labels := metrics.ExtractLabels(c, req.Model)
	metrics.TokensConsumedTotal.WithLabelValues(labels...).Add(float64(promptTokens))

	resp := model.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  req.Model,
		Usage: model.Usage{
			PromptTokens:     promptTokens,
			TotalTokens:      promptTokens,
			CompletionTokens: 0,
		},
	}

	return c.JSON(http.StatusOK, resp)
}
