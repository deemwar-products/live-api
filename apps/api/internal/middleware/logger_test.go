package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalWriter)

	r := gin.New()
	r.Use(Logger())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.Code)
	}

	logged := buf.String()
	for _, part := range []string{"GET", "/ping", "204"} {
		if !strings.Contains(logged, part) {
			t.Fatalf("expected log %q to contain %q", logged, part)
		}
	}
}
