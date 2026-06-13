package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestExtractLabels(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Org-Id", "org123")
	req.Header.Set("X-Project-Id", "proj456")
	req.Header.Set("X-User-Id", "user789")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	labels := ExtractLabels(c, "test-model")
	assert.Equal(t, []string{"org123", "proj456", "user789", "test-model"}, labels)
}

func TestExtractLabels_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	labels := ExtractLabels(c, "")
	assert.Equal(t, []string{"unknown", "unknown", "unknown", "unknown"}, labels)
}

func TestPrometheusMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := PrometheusMiddleware()
	handler := middleware(func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
