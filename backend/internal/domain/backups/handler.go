package backups

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
)

type Handler struct {
	scheduler *Scheduler
}

func NewHandler(scheduler *Scheduler) *Handler {
	return &Handler{scheduler: scheduler}
}

type BackupListResponse struct {
	Items      []BackupFile `json:"items"`
	Status     string       `json:"status"`
	LastRunAt  *time.Time   `json:"last_run_at,omitempty"`
	LastStatus string       `json:"last_status,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	backups := r.Group("/backups")
	backups.Use(middleware.RequireAnyPermission("observability:view_health", "users:view"))
	{
		backups.GET("", h.ListBackups)
		backups.POST("/run", middleware.RequirePermission("observability:manage_alerts"), h.TriggerBackup)
	}
}

func (h *Handler) ListBackups(c *gin.Context) {
	last := h.scheduler.LastRun()
	resp := BackupListResponse{
		Items:  h.scheduler.History(),
		Status: string(h.scheduler.Status()),
	}
	if last != nil {
		resp.LastRunAt = &last.CreatedAt
		resp.LastStatus = string(last.Status)
	}
	response.OK(c, resp)
}

func (h *Handler) TriggerBackup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	if err := h.scheduler.RunBackup(ctx); err != nil {
		if err.Error() == "backup already in progress" {
			response.Conflict(c, "BACKUP_IN_PROGRESS", "backup already in progress")
			return
		}
		response.InternalServerError(c)
		_ = c.Error(err)
		return
	}

	last := h.scheduler.LastRun()
	resp := BackupListResponse{
		Items:      h.scheduler.History(),
		Status:     string(h.scheduler.Status()),
		LastStatus: string(last.Status),
	}
	if last != nil {
		resp.LastRunAt = &last.CreatedAt
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}
