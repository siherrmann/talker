# talker

[![Go Reference](https://pkg.go.dev/badge/github.com/siherrmann/talker.svg)](https://pkg.go.dev/github.com/siherrmann/talker)
[![Go Report Card](https://goreportcard.com/badge/siherrmann/talker)](http://goreportcard.com/report/siherrmann/talker)

A fast, OpenAI-compatible Chat Completion API wrapping local LLM inference using [hugot](https://github.com/knights-analytics/hugot).

---

## 💡 Goal of this project

talker provides a lightweight, entirely local backend that mimics the OpenAI Chat Completion API (`POST /v1/chat/completions`) and Embeddings API (`POST /v1/embeddings`). It enables you to point your existing OpenAI-compatible AI applications directly to a local, privacy-preserving server running ONNX-based language models without needing complex Python setups.

---

## 🛠️ Installation

To run talker, clone the repository and run it via Go:

```bash
git clone https://github.com/siherrmann/talker.git
cd talker
go mod tidy
```

The server requires:

- Go 1.25+
- ONNX-formatted language or embedding models (which can be downloaded automatically!)

---

## 🚀 Getting Started

### Basic Usage

The simplest way to start the API for testing is by using the built-in mock engine. If no model parameters are specified, the server will default to the mock engine, allowing you to test endpoints immediately.

```bash
go run main.go
```

To run with real models and have them download automatically if they are missing:

```bash
MODEL_FOLDER=./models CHAT_MODEL=HuggingFaceTB/SmolLM-135M-Instruct EMBEDDING_MODEL=BAAI/bge-small-en-v1.5 PORT=8080 go run main.go
```

### Environment Variables

The API behavior can be configured via environment variables:

```shell
MODEL_FOLDER=./models                        # Required for auto-download: The base directory to store models.
CHAT_MODEL=HuggingFaceTB/SmolLM-135M-Instruct # Optional: The Hugging Face repo name for the text generation model.
EMBEDDING_MODEL=BAAI/bge-small-en-v1.5        # Optional: The Hugging Face repo name for the embeddings model.
PORT=8080                                    # Optional: Sets the port for the Echo server (default is 8080).
```

If neither `CHAT_MODEL` nor `EMBEDDING_MODEL` is provided, the mock engine is used.

---

## ⭐ Features

### Local LLM Inference

- **hugot Integration**: Native Go inference using the high-performance `hugot` library (which wraps ONNX runtime).
- **Automatic Downloading**: Automatically downloads the requested models from Hugging Face directly into your `MODEL_FOLDER` on startup.

### OpenAI Compatibility

- **Standard Endpoints**: Strict implementation of both `POST /v1/chat/completions` and `POST /v1/embeddings`.
- **Request/Response Models**: Fully conforms to the standard OpenAI request and response schemas.
- **SSE Streaming**: Fully supports Server-Sent Events for real-time streaming when `stream: true` is passed.
- **Strict JSON Enforcement**: Supports `response_format: {"type": "json_object"}` with automatic struct validation via `github.com/siherrmann/validator`. If the LLM generates invalid JSON, the engine automatically retries up to 3 times, passing the validation errors back to the model as a prompt.

### Robust Architecture

- **Echo v5 Framework**: Built on top of Echo for rapid and robust HTTP routing.
- **Test-Driven**: Designed with a highly mockable architecture.

---

## 🖥️ API Interface

### API Endpoints

- **`POST /v1/chat/completions`** - Generates chat completions.
- **`POST /v1/embeddings`** - Generates vector embeddings for a given input.

Example request (Non-streaming Chat):

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "local-model",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

Example request (Embeddings):

```bash
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "local-embedding-model",
    "input": ["First sentence", "Second sentence"]
  }'
```

---

## 🏗️ Architecture

talker is built with:

- **[Echo v5](https://echo.labstack.com/)** - Fast HTTP framework for Go
- **[hugot](https://github.com/knights-analytics/hugot)** - Golang wrapper around ONNX Runtime for local inference pipelines

The application follows a clean architecture with:

- **Handlers (`handler/`)**: Contains `ChatHandler` and `EmbeddingsHandler` for the HTTP lifecycle.
- **Core Engine (`core/`)**: Abstracts underlying `hugot` pipeline calls (`HugotEngine`). It seamlessly supports `TextGenerationPipeline` and `FeatureExtractionPipeline` concurrently.
- **Models (`model/`)**: Native Go structs matching the exact schema required by client libraries expecting an OpenAI backend. Includes custom unmarshaling logic for robust handling of dynamic OpenAI fields (e.g., embeddings input as string vs array).

---

## 🔧 Development

### Prerequisites

- Go 1.25+

### Development Commands

Run the test suite to verify handlers and data parsing logic:

```bash
# Run all tests
go test ./...

# Run server with Mock Engine
go run main.go
```
