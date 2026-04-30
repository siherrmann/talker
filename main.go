package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/siherrmann/talker/core"
	"github.com/siherrmann/talker/handler"
)

func main() {
	// 1. Initialize the Engine
	modelFolder := os.Getenv("MODEL_FOLDER")
	speechModel := os.Getenv("CHAT_MODEL")
	embeddingModel := os.Getenv("EMBEDDING_MODEL")

	var engine core.Engine
	var err error

	if modelFolder == "" || (speechModel == "" && embeddingModel == "") {
		log.Println("WARNING: MODEL_FOLDER or model names not fully set, using MockEngine for testing.")
		engine = &core.MockEngine{}
	} else {
		log.Printf("Initializing Hugot engine with speech model: %s, embedding model: %s in folder: %s", speechModel, embeddingModel, modelFolder)
		engine, err = core.NewHugotEngine(modelFolder, speechModel, embeddingModel)
		if err != nil {
			log.Fatalf("Failed to initialize Hugot Engine: %v", err)
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

	log.Printf("Starting Talker API on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
