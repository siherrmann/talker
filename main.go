package main

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/siherrmann/talker/core"
	"github.com/siherrmann/talker/handler"
)

func run() error {
	// 1. Initialize the Engine
	modelFolder := os.Getenv("MODEL_FOLDER")
	speechModel := os.Getenv("CHAT_MODEL")
	embeddingModel := os.Getenv("EMBEDDING_MODEL")
	
	var engine core.Engine
	var err error

	if modelFolder == "" || (speechModel == "" && embeddingModel == "") {
		slog.Warn("MODEL_FOLDER or model names not fully set, using MockEngine for testing.")
		engine = &core.MockEngine{}
	} else {
		slog.Info("Initializing Hugot engine", "speech_model", speechModel, "embedding_model", embeddingModel, "folder", modelFolder)
		engine, err = core.NewHugotEngine(modelFolder, speechModel, embeddingModel)
		if err != nil {
			return err
		}
	}
	defer engine.Close()

	// 2. Initialize Handlers
	chatHandler := handler.NewChatHandler(engine)
	embeddingsHandler := handler.NewEmbeddingsHandler(engine)

	// 3. Setup Echo Server
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())

	// Register Routes
	e.POST("/v1/chat/completions", chatHandler.ChatCompletions)
	e.POST("/v1/embeddings", embeddingsHandler.Embeddings)

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Starting Talker API", "port", port)
	return e.Start(":" + port)
}

func main() {
	if err := run(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
