package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/siherrmann/talker/handler"
	"github.com/siherrmann/talker/llm"
)

func main() {
	// 1. Initialize the Engine
	modelPath := os.Getenv("MODEL_PATH")
	var engine llm.Engine
	var err error

	if modelPath == "" {
		log.Println("WARNING: MODEL_PATH not set, using MockEngine for testing.")
		engine = &llm.MockEngine{}
	} else {
		log.Printf("Initializing Hugot engine with model: %s", modelPath)
		engine, err = llm.NewHugotEngine(modelPath)
		if err != nil {
			log.Fatalf("Failed to initialize Hugot Engine: %v", err)
		}
	}
	defer engine.Close()

	// 2. Initialize TalkerHandler
	talkerHandler := handler.NewTalkerHandler(engine)

	// 3. Setup Echo Server
	e := echo.New()
	
	// Middleware
	e.Use(middleware.Recover())

	// Register Routes
	talkerHandler.RegisterRoutes(e)

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
