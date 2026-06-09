package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/deemwar/live-api-poc/apps/api/internal/service"
)

type JobServicer interface {
	CreateJob(ctx context.Context, orgID, jobType string) (*service.Job, error)
	ListJobs(ctx context.Context) ([]service.Job, error)
	GetJob(ctx context.Context, id string) (*service.Job, error)
	CancelJob(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*service.JobStats, error)
}

type JobHandler struct {
	svc JobServicer
}

func NewJobHandler(svc JobServicer) *JobHandler {
	return &JobHandler{svc: svc}
}

type createJobRequest struct {
	OrgID string `json:"org_id" binding:"required"`
	Type  string `json:"type"   binding:"required"`
}

func (h *JobHandler) List(c *gin.Context) {
	jobs, err := h.svc.ListJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *JobHandler) Get(c *gin.Context) {
	job, err := h.svc.GetJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) Create(c *gin.Context) {
	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.svc.CreateJob(c.Request.Context(), req.OrgID, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) Cancel(c *gin.Context) {
	if err := h.svc.CancelJob(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func (h *JobHandler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
