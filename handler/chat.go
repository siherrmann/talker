package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/core"
	"github.com/siherrmann/talker/model"
	"github.com/siherrmann/validator"
)

type ChatHandler struct {
	Engine    core.Engine
	validator *validator.Validator
}

func NewChatHandler(engine core.Engine) *ChatHandler {
	return &ChatHandler{
		Engine:    engine,
		validator: validator.NewValidator(),
	}
}

func (h *ChatHandler) ChatCompletions(c *echo.Context) error {
	var req model.ChatCompletionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if len(req.Messages) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Messages array cannot be empty"})
	}

	// Format messages into a simple prompt template
	// In a real scenario, you'd use a specific chat template like ChatML
	var promptBuilder strings.Builder
	for _, msg := range req.Messages {
		promptBuilder.WriteString(msg.Role + ": " + msg.Content + "\n")
	}
	promptBuilder.WriteString("assistant: ")

	prompt := promptBuilder.String()

	if req.Stream {
		tokenChan, errChan, err := h.Engine.GenerateStream(c.Request().Context(), prompt, req.MaxTokens)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate completion stream: " + err.Error()})
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")

		id := "chatcmpl-" + uuid.New().String()
		created := time.Now().Unix()

		for {
			select {
			case <-c.Request().Context().Done():
				return nil
			case err, ok := <-errChan:
				if ok && err != nil {
					// We can't really change the HTTP status code now, but we can send an error message
					return err
				}
			case token, ok := <-tokenChan:
				if !ok {
					// Stream finished
					stopReason := "stop"
					chunk := model.ChatCompletionChunkResponse{
						ID:      id,
						Object:  "chat.completion.chunk",
						Created: created,
						Model:   req.Model,
						Choices: []model.ChunkChoice{
							{
								Index:        0,
								Delta:        model.ChunkDelta{},
								FinishReason: &stopReason,
							},
						},
					}
					if err := streamChunk(c, chunk); err != nil {
						return err
					}
					c.Response().Write([]byte("data: [DONE]\n\n"))
					if flusher, ok := c.Response().(http.Flusher); ok {
						flusher.Flush()
					}
					return nil
				}

				chunk := model.ChatCompletionChunkResponse{
					ID:      id,
					Object:  "chat.completion.chunk",
					Created: created,
					Model:   req.Model,
					Choices: []model.ChunkChoice{
						{
							Index: 0,
							Delta: model.ChunkDelta{
								Content: token,
							},
						},
					},
				}
				if err := streamChunk(c, chunk); err != nil {
					return err
				}
			}
		}
	}

	// Non-streaming
	var output string
	var err error
	maxRetries := 3

	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_object" {
		currentPrompt := prompt
		for i := 0; i < maxRetries; i++ {
			output, err = h.Engine.Generate(c.Request().Context(), currentPrompt, req.MaxTokens)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate completion: " + err.Error()})
			}

			err = core.ValidateJSON(output, h.validator)
			if err == nil {
				break
			}

			if i == maxRetries-1 {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate valid JSON after multiple attempts: " + err.Error()})
			}

			currentPrompt += output + "\nThe JSON you provided was invalid. Error: " + err.Error() + ". Please try again and provide ONLY valid JSON.\nassistant: "
		}
	} else {
		output, err = h.Engine.Generate(c.Request().Context(), prompt, req.MaxTokens)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate completion: " + err.Error()})
		}
	}

	promptTokens := len(strings.Fields(prompt))
	completionTokens := len(strings.Fields(output))

	resp := model.ChatCompletionResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []model.Choice{
			{
				Index: 0,
				Message: model.ChatMessage{
					Role:    "assistant",
					Content: output,
				},
				FinishReason: "stop",
			},
		},
		Usage: model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func streamChunk(c *echo.Context, chunk model.ChatCompletionChunkResponse) error {
	b, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	if _, err := c.Response().Write([]byte("data: " + string(b) + "\n\n")); err != nil {
		return err
	}
	if flusher, ok := c.Response().(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
