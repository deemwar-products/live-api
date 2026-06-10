package server

import (
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/gin-gonic/gin"
)

func init() {
 gin.SetMode(gin.TestMode)
}

func TestServer_WhenNewCalled_ThenReturnsServer(t *testing.T) {
 srv := New()
 if srv == nil {
 t.Fatal("New returned nil")
 }
 if srv.Engine() == nil {
 t.Error("Engine not initialized")
 }
 if srv.address == "" {
 t.Error("Address not set")
 }
}

func TestServer_WhenHealthEndpointHit_ThenReturnsOk(t *testing.T) {
 srv := New()
 srv.addMiddleware()
 srv.registerRoutes()

 req, _ := http.NewRequest("GET", "/health", nil)
 w := httptest.NewRecorder()
 srv.Engine().ServeHTTP(w, req)

 if w.Code != http.StatusOK {
 t.Errorf("Expected status 200, got %d", w.Code)
 }

 expected := `{"status":"ok"}`
 if w.Body.String() != expected {
 t.Errorf("Expected body %s, got %s", expected, w.Body.String())
 }
}

func TestServer_WhenHealthEndpointHit_ThenContentTypeIsJSON(t *testing.T) {
 srv := New()
 srv.addMiddleware()
 srv.registerRoutes()

 req, _ := http.NewRequest("GET", "/health", nil)
 w := httptest.NewRecorder()
 srv.Engine().ServeHTTP(w, req)

 contentType := w.Header().Get("Content-Type")
 if contentType != "application/json; charset=utf-8" {
 t.Errorf("Expected application/json, got %s", contentType)
 }
}

func TestServer_WhenAddMiddlewareCalled_ThenRecoveryIsRegistered(t *testing.T) {
 srv := New()
 srv.addMiddleware()

 handlers := srv.Engine().Handlers
 if len(handlers) == 0 {
 t.Error("No middleware registered")
 }
}

func TestServer_WhenRegisterRoutesCalled_ThenHealthRouteExists(t *testing.T) {
 srv := New()
 srv.registerRoutes()

 routes := srv.Engine().Routes()
 found := false
 for _, route := range routes {
 if route.Path == "/health" && route.Method == "GET" {
 found = true
 break
 }
 }
 if !found {
 t.Error("Health route not registered")
 }
}