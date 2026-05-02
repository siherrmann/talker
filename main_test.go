package main

import (
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunFunc(t *testing.T) {
	// Setup environment
	os.Setenv("PORT", "18081") // Use a specific port to test
	os.Setenv("MODEL_FOLDER", "") // Forces mock engine

	// Run the app in a goroutine
	go func() {
		_ = run()
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify the server is running by making a request
	resp, err := http.Post("http://localhost:18081/v1/chat/completions", "application/json", nil)
	assert.NoError(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRunFunc_DefaultPort(t *testing.T) {
	// Setup environment
	os.Setenv("PORT", "") // Force default port 8080
	os.Setenv("MODEL_FOLDER", "") // Forces mock engine

	// Run the app in a goroutine
	go func() {
		_ = run()
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify the server is running on 8080
	resp, err := http.Post("http://localhost:8080/v1/chat/completions", "application/json", nil)
	assert.NoError(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRunFunc_Error(t *testing.T) {
	// Setup environment to force Hugot error by giving an invalid model folder
	os.Setenv("MODEL_FOLDER", "/root/invalid_permissions")
	os.Setenv("CHAT_MODEL", "test-model")
	
	err := run()
	assert.Error(t, err)
}

func TestMain_Exit(t *testing.T) {
	if os.Getenv("TEST_MAIN_EXIT") == "1" {
		os.Setenv("MODEL_FOLDER", "/root/forbidden_dir_test")
		os.Setenv("CHAT_MODEL", "test")
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_Exit")
	cmd.Env = append(os.Environ(), "TEST_MAIN_EXIT=1")
	err := cmd.Run()
	
	assert.Error(t, err)
}
