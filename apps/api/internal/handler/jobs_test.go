package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/deemwar/live-api-poc/apps/api/internal/service"
)

type fakeJobService struct {
	createJobFn func(context.Context, string, string) (*service.Job, error)
	listJobsFn  func(context.Context) ([]service.Job, error)
	getJobFn    func(context.Context, string) (*service.Job, error)
	cancelJobFn func(context.Context, string) error
	getStatsFn  func(context.Context) (*service.JobStats, error)
}

func (f *fakeJobService) CreateJob(ctx context.Context, orgID, jobType string) (*service.Job, error) {
	if f.createJobFn == nil {
		return nil, nil
	}
	return f.createJobFn(ctx, orgID, jobType)
}

func (f *fakeJobService) ListJobs(ctx context.Context) ([]service.Job, error) {
	if f.listJobsFn == nil {
		return nil, nil
	}
	return f.listJobsFn(ctx)
}

func (f *fakeJobService) GetJob(ctx context.Context, id string) (*service.Job, error) {
	if f.getJobFn == nil {
		return nil, nil
	}
	return f.getJobFn(ctx, id)
}

func (f *fakeJobService) CancelJob(ctx context.Context, id string) error {
	if f.cancelJobFn == nil {
		return nil
	}
	return f.cancelJobFn(ctx, id)
}

func (f *fakeJobService) GetStats(ctx context.Context) (*service.JobStats, error) {
	if f.getStatsFn == nil {
		return nil, nil
	}
	return f.getStatsFn(ctx)
}

func testRouter(svc JobServicer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewJobHandler(svc)
	r := gin.New()
	r.GET("/jobs", h.List)
	r.GET("/jobs/:id", h.Get)
	r.POST("/jobs", h.Create)
	r.DELETE("/jobs/:id", h.Cancel)
	r.GET("/stats", h.Stats)
	return r
}

func TestJobHandlerList(t *testing.T) {
	r := testRouter(&fakeJobService{listJobsFn: func(context.Context) ([]service.Job, error) {
		return []service.Job{{ID: "job_1", Type: "type", OrgID: "org", Status: "queued", CreatedAt: time.Now().UTC()}}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestJobHandlerListError(t *testing.T) {
	r := testRouter(&fakeJobService{listJobsFn: func(context.Context) ([]service.Job, error) {
		return nil, errors.New("boom")
	}})

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestJobHandlerGet(t *testing.T) {
	r := testRouter(&fakeJobService{getJobFn: func(context.Context, string) (*service.Job, error) {
		job := &service.Job{ID: "job_1", Type: "type", OrgID: "org", Status: "queued", CreatedAt: time.Now().UTC()}
		return job, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/jobs/job_1", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestJobHandlerGetNotFound(t *testing.T) {
	r := testRouter(&fakeJobService{getJobFn: func(context.Context, string) (*service.Job, error) {
		return nil, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/jobs/missing", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestJobHandlerGetError(t *testing.T) {
	r := testRouter(&fakeJobService{getJobFn: func(context.Context, string) (*service.Job, error) {
		return nil, errors.New("boom")
	}})

	req := httptest.NewRequest(http.MethodGet, "/jobs/job_1", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestJobHandlerCreate(t *testing.T) {
	r := testRouter(&fakeJobService{createJobFn: func(_ context.Context, orgID, jobType string) (*service.Job, error) {
		return &service.Job{ID: "job_1", OrgID: orgID, Type: jobType, Status: "queued", CreatedAt: time.Now().UTC()}, nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"org_id":"org","type":"process_document"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}
}

func TestJobHandlerCreateBadRequest(t *testing.T) {
	r := testRouter(&fakeJobService{})

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"org_id":"org"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestJobHandlerCreateError(t *testing.T) {
	r := testRouter(&fakeJobService{createJobFn: func(context.Context, string, string) (*service.Job, error) {
		return nil, errors.New("boom")
	}})

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"org_id":"org","type":"process_document"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestJobHandlerCancel(t *testing.T) {
	r := testRouter(&fakeJobService{cancelJobFn: func(context.Context, string) error { return nil }})

	req := httptest.NewRequest(http.MethodDelete, "/jobs/job_1", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestJobHandlerCancelError(t *testing.T) {
	r := testRouter(&fakeJobService{cancelJobFn: func(context.Context, string) error { return errors.New("boom") }})

	req := httptest.NewRequest(http.MethodDelete, "/jobs/job_1", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestJobHandlerStats(t *testing.T) {
	r := testRouter(&fakeJobService{getStatsFn: func(context.Context) (*service.JobStats, error) {
		return &service.JobStats{Queued: 1}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var body service.JobStats
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Queued != 1 {
		t.Fatalf("unexpected stats: %+v", body)
	}
}

func TestJobHandlerStatsError(t *testing.T) {
	r := testRouter(&fakeJobService{getStatsFn: func(context.Context) (*service.JobStats, error) {
		return nil, errors.New("boom")
	}})

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}
