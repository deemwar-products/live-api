package server

import (
 "database/sql"
 "fmt"
 "net"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"

 "github.com/deemwar/live-api/apps/api/internal/config"
 "github.com/deemwar/live-api/apps/api/internal/db"
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

func TestServer_WhenDbOpenFails_ThenPanics(t *testing.T) {
 t.Setenv("DB_PATH", "/tmp/whatever.db")
 config.Reset()

 badOpener := func(string) (*sql.DB, error) {
 return nil, fmt.Errorf("simulated open failure")
 }

 defer func() {
 if r := recover(); r == nil {
 t.Error("expected panic when db open fails")
 }
 }()

 _ = NewWithDeps(badOpener, func(conn *sql.DB) error {
 	return db.Migrate(conn, t.TempDir())
 })
}

func TestServer_WhenMigrateFails_ThenStartReturnsError(t *testing.T) {
 port := freePort(t)
 t.Setenv("PORT", fmt.Sprintf("%d", port))
 t.Setenv("DB_PATH", t.TempDir()+"/test.db")
 config.Reset()

 badMigrator := func(*sql.DB) error {
 return fmt.Errorf("simulated migrate failure")
 }

 srv := NewWithDeps(db.OpenReadWrite, badMigrator)
 if err := srv.Start(); err == nil {
 t.Error("expected migrate error, got nil")
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

func TestServer_WhenStartCalled_ThenServesRequests(t *testing.T) {
 port := freePort(t)
 t.Setenv("PORT", fmt.Sprintf("%d", port))
 t.Setenv("DB_PATH", t.TempDir()+"/test.db")
 t.Setenv("MIGRATIONS_PATH", "../../../migrations")
 config.Reset()

 srv := New()

 done := make(chan struct{})
 go func() {
 defer close(done)
 _ = srv.Start()
 }()

 waitForReady(t, fmt.Sprintf("http://127.0.0.1:%d/health", port))

 resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
 if err != nil {
 t.Fatalf("health request failed: %v", err)
 }
 defer resp.Body.Close()

 if resp.StatusCode != http.StatusOK {
 t.Errorf("expected 200, got %d", resp.StatusCode)
 }
}

func freePort(t *testing.T) int {
 t.Helper()
 listener, err := net.Listen("tcp", "127.0.0.1:0")
 if err != nil {
 t.Fatalf("could not allocate port: %v", err)
 }
 defer listener.Close()
 return listener.Addr().(*net.TCPAddr).Port
}

func waitForReady(t *testing.T, url string) {
 t.Helper()
 deadline := time.Now().Add(3 * time.Second)
 for time.Now().Before(deadline) {
 resp, err := http.Get(url)
 if err == nil {
 resp.Body.Close()
 if resp.StatusCode == http.StatusOK {
 return
 }
 }
 time.Sleep(20 * time.Millisecond)
 }
 t.Fatalf("server did not become ready at %s", url)
}
