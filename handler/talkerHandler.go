package handler

import (
	"github.com/labstack/echo/v5"
	"github.com/siherrmann/talker/llm"
)

type TalkerHandler struct {
	Engine llm.Engine
}

func NewTalkerHandler(engine llm.Engine) *TalkerHandler {
	return &TalkerHandler{
		Engine: engine,
	}
}

func (h *TalkerHandler) RegisterRoutes(e *echo.Echo) {
	v1 := e.Group("/v1")
	v1.POST("/chat/completions", h.ChatCompletions)
}
