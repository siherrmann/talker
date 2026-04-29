# talker

[![Go Reference](https://pkg.go.dev/badge/github.com/siherrmann/talker.svg)](https://pkg.go.dev/github.com/siherrmann/talker)
[![Go Report Card](https://goreportcard.com/badge/siherrmann/talker)](http://goreportcard.com/report/siherrmann/talker)

A fast, OpenAI-compatible Chat Completion API wrapping local LLM inference using [hugot](https://github.com/knights-analytics/hugot).

---

## 💡 Goal of this project

talker provides a lightweight, entirely local backend that mimics the OpenAI Chat Completion API (`POST /v1/chat/completions`). It enables you to point your existing OpenAI-compatible AI applications directly to a local, privacy-preserving server running ONNX-based language models without needing complex Python setups.

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
- An ONNX-formatted language model

---

## 🚀 Getting Started

### Basic Usage

The simplest way to start the API for testing is by using the built-in mock engine. If no model path is specified, the server will default to the mock engine, allowing you to test endpoints immediately.

```bash
go run main.go
```

To run with a real LLM model for inference:

```bash
MODEL_PATH=/path/to/your/model.onnx PORT=8080 go run main.go
```

### Environment Variables

The API behavior can be configured via environment variables:

```shell
MODEL_PATH=/path/to/model.onnx # Optional: Sets the ONNX model directory. If empty, MockEngine is used.
PORT=8080                      # Optional: Sets the port for the Echo server (default is 8080).
```

---

## ⭐ Features

### Local LLM Inference

- **hugot Integration**: Native Go inference using the high-performance `hugot` library (which wraps ONNX runtime).
- **Direct Downloads**: Includes wrappers to easily download models directly from the Hugging Face Hub.

### OpenAI Compatibility

- **Standard Endpoints**: Provides a strict implementation of the `POST /v1/chat/completions` endpoint.
- **Request/Response Models**: Fully conforms to the standard OpenAI request (messages, temperature, max_tokens, etc.) and response structures.
- **SSE Streaming**: Fully supports Server-Sent Events for real-time streaming when `stream: true` is passed. 

### Robust Architecture

- **Echo v5 Framework**: Built on top of Echo for rapid and robust HTTP routing.
- **Test-Driven**: Designed with a highly mockable architecture heavily mirroring the `queuerManager` style, including extensive unit tests.

---

## 🖥️ API Interface

The server provides a single, OpenAI-compatible completion endpoint.

### API Endpoints

- **`POST /v1/chat/completions`** - Generates chat completions.

Example request (Non-streaming):

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

Example request (Streaming):

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "local-model",
    "messages": [{"role": "user", "content": "Tell me a joke."}],
    "stream": true
  }'
```

---

## 🏗️ Architecture

talker is built with:

- **[Echo v5](https://echo.labstack.com/)** - Fast HTTP framework for Go
- **[hugot](https://github.com/knights-analytics/hugot)** - Golang wrapper around ONNX Runtime for local inference pipelines

The application follows a clean architecture with:

- **Handlers**: The Echo `TalkerHandler` handles HTTP lifecycle and parses OpenAI JSON structures.
- **LLM Engine**: A mockable `Engine` interface that abstracts underlying `hugot` pipeline calls (`HugotEngine`), generating either static string responses or continuous token streams.
- **Models**: Native Go structs matching the exact schema required by client libraries expecting an OpenAI backend.

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
