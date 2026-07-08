package auditlogs

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	audit := r.Group("/audit-logs")
	audit.Use(middleware.RequirePermission("observability:view_audit"))
	{
		audit.GET("", h.List)
		audit.GET("/export", h.Export)
	}
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := h.parseFilters(c)

	logs, total, err := h.store.ListAuditLogs(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": logs,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Export(c *gin.Context) {
	filters := h.parseFilters(c)

	logs, err := h.store.ListAuditLogsAll(c.Request.Context(), filters)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID", "Timestamp", "Action", "ActorID", "ActorRole", "EntityType", "EntityID", "LocationID", "IPAddress"})

	for _, l := range logs {
		actorID := ""
		if l.ActorID != nil {
			actorID = *l.ActorID
		}
		actorRole := ""
		if l.ActorRole != nil {
			actorRole = *l.ActorRole
		}
		locationID := ""
		if l.LocationID != nil {
			locationID = *l.LocationID
		}
		ip := ""
		if l.IPAddress != nil {
			ip = *l.IPAddress
		}

		writer.Write([]string{
			l.ID,
			l.Timestamp.Format(time.RFC3339),
			l.Action,
			actorID,
			actorRole,
			l.EntityType,
			l.EntityID,
			locationID,
			ip,
		})
	}
	writer.Flush()
}

func (h *Handler) parseFilters(c *gin.Context) store.ListAuditLogsFilters {
	filters := store.ListAuditLogsFilters{
		Action:     c.Query("action"),
		ActorID:    c.Query("actor_id"),
		EntityType: c.Query("entity_type"),
		EntityID:   c.Query("entity_id"),
		LocationID: c.Query("location_id"),
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.DateFrom = &t
		} else {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				filters.DateFrom = &t
			}
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.DateTo = &t
		} else {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				filters.DateTo = &t
			}
		}
	}
	fmt.Printf("parsed filters: %+v\n", filters)
	return filters
}