package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/deemwar/live-api-poc/apps/api/internal/config"
	"github.com/deemwar/live-api-poc/apps/api/internal/handler"
	"github.com/deemwar/live-api-poc/apps/api/internal/service"
)

type stubJobService struct{}

func (stubJobService) CreateJob(context.Context, string, string) (*service.Job, error) {
	return &service.Job{ID: "job_1", CreatedAt: time.Now().UTC()}, nil
}

func (stubJobService) ListJobs(context.Context) ([]service.Job, error) { return nil, nil }

func (stubJobService) GetJob(context.Context, string) (*service.Job, error) {
	return &service.Job{ID: "job_1", CreatedAt: time.Now().UTC()}, nil
}

func (stubJobService) CancelJob(context.Context, string) error { return nil }

func (stubJobService) GetStats(context.Context) (*service.JobStats, error) {
	return &service.JobStats{Queued: 1}, nil
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func TestNewRouterRegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouter(handler.NewJobHandler(stubJobService{}))

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{method: http.MethodGet, path: "/health", code: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/queue/jobs", code: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/queue/jobs/job_1", code: http.StatusOK},
		{method: http.MethodDelete, path: "/api/v1/queue/jobs/job_1", code: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/queue/stats", code: http.StatusOK},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != tc.code {
			t.Fatalf("%s %s: expected %d, got %d", tc.method, tc.path, tc.code, res.Code)
		}
	}
}

func TestRun(t *testing.T) {
	origLoadConfig := loadConfig
	origOpenDatabase := openDatabase
	origNewRedis := newRedis
	origRunHTTPServer := runHTTPServer
	origLogPrintf := logPrintf
	defer func() {
		loadConfig = origLoadConfig
		openDatabase = origOpenDatabase
		newRedis = origNewRedis
		runHTTPServer = origRunHTTPServer
		logPrintf = origLogPrintf
	}()

	loadConfig = func() *config.Config {
		return &config.Config{Port: "8080", RedisHost: "redis", RedisPort: "6379", DBPath: "memory"}
	}
	openDatabase = func(path string) (io.Closer, error) {
		if path != "memory" {
			t.Fatalf("unexpected db path %q", path)
		}
		return nopCloser{}, nil
	}
	newRedis = func(addr string) *redis.Client {
		if addr != "redis:6379" {
			t.Fatalf("unexpected redis addr %q", addr)
		}
		return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	}
	logPrintf = func(string, ...any) {}
	runHTTPServer = func(r *gin.Engine, addr string) error {
		if addr != ":8080" {
			t.Fatalf("unexpected listen addr %q", addr)
		}
		return nil
	}

	if err := run(); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunOpenDatabaseError(t *testing.T) {
	origLoadConfig := loadConfig
	origOpenDatabase := openDatabase
	defer func() {
		loadConfig = origLoadConfig
		openDatabase = origOpenDatabase
	}()

	loadConfig = func() *config.Config { return &config.Config{DBPath: "bad"} }
	openDatabase = func(string) (io.Closer, error) { return nil, errors.New("boom") }

	if err := run(); err == nil {
		t.Fatal("expected run error")
	}
}

func TestMainFatalPath(t *testing.T) {
	origRunHTTPServer := runHTTPServer
	origOpenDatabase := openDatabase
	origLoadConfig := loadConfig
	origNewRedis := newRedis
	origLogPrintf := logPrintf
	origLogFatalf := logFatalf
	defer func() {
		runHTTPServer = origRunHTTPServer
		openDatabase = origOpenDatabase
		loadConfig = origLoadConfig
		newRedis = origNewRedis
		logPrintf = origLogPrintf
		logFatalf = origLogFatalf
	}()

	loadConfig = func() *config.Config {
		return &config.Config{Port: "8080", RedisHost: "redis", RedisPort: "6379", DBPath: "memory"}
	}
	openDatabase = func(string) (io.Closer, error) { return nopCloser{}, nil }
	newRedis = func(string) *redis.Client { return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) }
	logPrintf = func(string, ...any) {}
	runHTTPServer = func(*gin.Engine, string) error { return errors.New("listen failed") }

	called := false
	logFatalf = func(string, ...any) {
		called = true
	}

	main()

	if !called {
		t.Fatal("expected main to call logFatalf")
	}
}
