package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

type Engine interface {
	Generate(ctx context.Context, prompt string, maxTokens int) (string, error)
	GenerateStream(ctx context.Context, prompt string, maxTokens int) (chan string, chan error, error)
	Close() error
}

func DownloadModel(ctx context.Context, modelName string, destination string) (string, error) {
	opts := hugot.NewDownloadOptions()
	return hugot.DownloadModel(ctx, modelName, destination, opts)
}

type HugotEngine struct {
	session  *hugot.Session
	pipeline *pipelines.TextGenerationPipeline
}

func NewHugotEngine(modelPath string) (*HugotEngine, error) {
	ctx := context.Background()
	session, err := hugot.NewORTSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create hugot session: %w", err)
	}

	config := hugot.TextGenerationConfig{
		ModelPath: modelPath,
	}

	pipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create text generation pipeline: %w", err)
	}

	return &HugotEngine{
		session:  session,
		pipeline: pipeline,
	}, nil
}

func (e *HugotEngine) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = 256
	}
	
	e.pipeline.MaxLength = maxTokens
	e.pipeline.Streaming = false

	result, err := e.pipeline.RunPipeline(ctx, []string{prompt})
	if err != nil {
		return "", err
	}

	if len(result.Responses) > 0 {
		return result.Responses[0], nil
	}
	
	return "", fmt.Errorf("no output generated")
}

func (e *HugotEngine) GenerateStream(ctx context.Context, prompt string, maxTokens int) (chan string, chan error, error) {
	if maxTokens <= 0 {
		maxTokens = 256
	}

	e.pipeline.MaxLength = maxTokens
	e.pipeline.Streaming = true

	result, err := e.pipeline.RunPipeline(ctx, []string{prompt})
	if err != nil {
		return nil, nil, err
	}

	tokenChan := make(chan string)
	
	go func() {
		defer close(tokenChan)
		for token := range result.TokenStream {
			tokenChan <- token.Token
		}
	}()

	return tokenChan, result.ErrorStream, nil
}

func (e *HugotEngine) Close() error {
	if e.session != nil {
		err := e.session.Destroy()
		e.session = nil
		return err
	}
	return nil
}

// MockEngine for testing purposes
type MockEngine struct {
	Response string
	Err      error
}

func (m *MockEngine) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Response == "" {
		return "This is a mock response to: " + strings.TrimSpace(prompt), nil
	}
	return m.Response, nil
}

func (m *MockEngine) GenerateStream(ctx context.Context, prompt string, maxTokens int) (chan string, chan error, error) {
	if m.Err != nil {
		return nil, nil, m.Err
	}
	
	tokenChan := make(chan string)
	errChan := make(chan error)
	
	go func() {
		defer close(tokenChan)
		defer close(errChan)
		
		response := m.Response
		if response == "" {
			response = "This is a mock response to: " + strings.TrimSpace(prompt)
		}
		
		words := strings.Split(response, " ")
		for i, word := range words {
			if i > 0 {
				tokenChan <- " "
			}
			tokenChan <- word
			time.Sleep(10 * time.Millisecond) // Simulate delay
		}
	}()
	
	return tokenChan, errChan, nil
}

func (m *MockEngine) Close() error {
	return nil
}
