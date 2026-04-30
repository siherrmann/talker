package handler

import (
	"log"
	"os"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/core"
)

var e *echo.Echo
var mockEngine *core.MockEngine
var chatHandler *ChatHandler
var embeddingsHandler *EmbeddingsHandler

func TestMain(m *testing.M) {
	mockEngine = &core.MockEngine{}
	chatHandler = NewChatHandler(mockEngine)
	embeddingsHandler = NewEmbeddingsHandler(mockEngine)

	e = echo.New()
	e.POST("/v1/chat/completions", chatHandler.ChatCompletions)
	e.POST("/v1/embeddings", embeddingsHandler.Embeddings)

	exitCode := m.Run()

	err := mockEngine.Close()
	if err != nil {
		log.Printf("Failed to close mock engine: %v", err)
	}

	os.Exit(exitCode)
}
