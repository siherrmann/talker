package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/backends"
	"github.com/knights-analytics/hugot/pipelines"
)

type Engine interface {
	Generate(ctx context.Context, prompt string, maxTokens int) (string, error)
	GenerateStream(ctx context.Context, prompt string, maxTokens int) (chan string, chan error, error)
	ExtractEmbeddings(ctx context.Context, input []string) ([][]float32, error)
	CountTokens(text string, isEmbedding bool) (int, error)
	Close() error
}

type HugotEngine struct {
	session           *hugot.Session
	pipeline          *pipelines.TextGenerationPipeline
	embeddingPipeline *pipelines.FeatureExtractionPipeline
}

func NewHugotEngine(modelFolder string, speechModel string, embeddingModel string) (*HugotEngine, error) {
	if err := os.MkdirAll(modelFolder, 0750); err != nil {
		return nil, fmt.Errorf("failed to create model folder: %w", err)
	}

	ctx := context.Background()
	opts := hugot.NewDownloadOptions()

	session, err := hugot.NewORTSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create hugot session: %w", err)
	}

	var chatPipeline *pipelines.TextGenerationPipeline
	if speechModel != "" {
		localPath := filepath.Join(modelFolder, speechModel)
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			slog.Info("Downloading model", "model", speechModel, "folder", modelFolder)
			if _, err := hugot.DownloadModel(ctx, speechModel, modelFolder, opts); err != nil {
				return nil, fmt.Errorf("failed to download speech model: %w", err)
			}
		}

		config := hugot.TextGenerationConfig{
			ModelPath: localPath,
		}
		chatPipeline, err = hugot.NewPipeline(session, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create text generation pipeline: %w", err)
		}
	}

	var embPipeline *pipelines.FeatureExtractionPipeline
	if embeddingModel != "" {
		localPath := filepath.Join(modelFolder, embeddingModel)
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			slog.Info("Downloading model", "model", embeddingModel, "folder", modelFolder)
			if _, err := hugot.DownloadModel(ctx, embeddingModel, modelFolder, opts); err != nil {
				return nil, fmt.Errorf("failed to download embedding model: %w", err)
			}
		}

		embConfig := hugot.FeatureExtractionConfig{
			ModelPath: localPath,
		}
		embPipeline, err = hugot.NewPipeline(session, embConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create feature extraction pipeline: %w", err)
		}
	}

	return &HugotEngine{
		session:           session,
		pipeline:          chatPipeline,
		embeddingPipeline: embPipeline,
	}, nil
}

func (e *HugotEngine) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if e.pipeline == nil {
		return "", fmt.Errorf("text generation pipeline not initialized")
	}

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
	if e.pipeline == nil {
		return nil, nil, fmt.Errorf("text generation pipeline not initialized")
	}

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

func (e *HugotEngine) ExtractEmbeddings(ctx context.Context, input []string) ([][]float32, error) {
	if e.embeddingPipeline == nil {
		return nil, fmt.Errorf("feature extraction pipeline not initialized")
	}

	result, err := e.embeddingPipeline.RunPipeline(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Embeddings, nil
}

func (e *HugotEngine) CountTokens(text string, isEmbedding bool) (int, error) {
	var tok *backends.Tokenizer
	if isEmbedding {
		if e.embeddingPipeline != nil && e.embeddingPipeline.GetModel() != nil {
			tok = e.embeddingPipeline.GetModel().Tokenizer
		}
	} else {
		if e.pipeline != nil && e.pipeline.GetModel() != nil {
			tok = e.pipeline.GetModel().Tokenizer
		}
	}

	if tok == nil || tok.GoTokenizer == nil || tok.GoTokenizer.Tokenizer == nil {
		// Fallback if tokenizer isn't loaded or isn't a GoTokenizer
		return len(strings.Fields(text)), nil
	}

	ids := tok.GoTokenizer.Tokenizer.Encode(text)
	return len(ids), nil
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
	Responses []string
	Err       error
	StreamErr error
}

func (m *MockEngine) CountTokens(text string, isEmbedding bool) (int, error) {
	return len(strings.Fields(text)), nil
}

func (m *MockEngine) ExtractEmbeddings(ctx context.Context, input []string) ([][]float32, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	res := make([][]float32, len(input))
	for i := range input {
		// Just a dummy embedding of length 3
		res[i] = []float32{0.1, 0.2, 0.3}
	}
	return res, nil
}

func (m *MockEngine) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}

	if len(m.Responses) > 0 {
		resp := m.Responses[0]
		if len(m.Responses) > 1 {
			m.Responses = m.Responses[1:]
		}
		return resp, nil
	}

	return "This is a mock response to: " + strings.TrimSpace(prompt), nil
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

		response := "This is a mock response to: " + strings.TrimSpace(prompt)
		if len(m.Responses) > 0 {
			response = m.Responses[0]
			if len(m.Responses) > 1 {
				m.Responses = m.Responses[1:]
			}
		}

		if m.StreamErr != nil {
			errChan <- m.StreamErr
			return
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
