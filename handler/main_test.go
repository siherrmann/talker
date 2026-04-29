package handler

import (
	"log"
	"os"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/llm"
)

var e *echo.Echo
var mockEngine *llm.MockEngine
var testHandler *TalkerHandler

func TestMain(m *testing.M) {
	mockEngine = &llm.MockEngine{}
	testHandler = NewTalkerHandler(mockEngine)
	
	e = echo.New()
	testHandler.RegisterRoutes(e)

	exitCode := m.Run()

	err := mockEngine.Close()
	if err != nil {
		log.Printf("Failed to close mock engine: %v", err)
	}

	os.Exit(exitCode)
}
