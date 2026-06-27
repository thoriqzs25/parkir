package incidents

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
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

type CreateIncidentRequest struct {
	LocationID  string  `json:"location_id" binding:"required,uuid"`
	Type        string  `json:"type" binding:"required,oneof=STUCK_AT_GATE PAYMENT_DISPUTE OPERATOR_ERROR SYSTEM_DOWNTIME"`
	SessionID   *string `json:"session_id,omitempty"`
	Description string  `json:"description" binding:"required"`
}

type ResolveIncidentRequest struct {
	ResolutionNotes string  `json:"resolution_notes" binding:"required"`
	AdjustmentAction *string `json:"adjustment_action,omitempty" binding:"omitempty,oneof=VOID_TRANSACTION REASSIGN_SESSION"`
	AdjustmentEntityID *string `json:"adjustment_entity_id,omitempty"`
	ManagerPIN       string  `json:"manager_pin,omitempty" binding:"omitempty,len=6"`
}

type AddNoteRequest struct {
	Note string `json:"note" binding:"required"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	incidents := r.Group("/incidents")
	incidents.Use(middleware.RequirePermission("incidents:view"))
	{
		incidents.GET("", h.List)
		incidents.GET("/:id", h.Get)
		incidents.GET("/:id/notes", h.ListNotes)
	}

	incidentsCreate := r.Group("/incidents")
	incidentsCreate.Use(middleware.RequirePermission("incidents:create"))
	{
		incidentsCreate.POST("", h.Create)
		incidentsCreate.POST("/:id/notes", h.AddNote)
	}

	incidentsResolve := r.Group("/incidents")
	incidentsResolve.Use(middleware.RequirePermission("incidents:resolve"))
	{
		incidentsResolve.PATCH("/:id/resolve", h.Resolve)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	inc, err := h.store.CreateIncident(c.Request.Context(), struct {
		LocationID  string
		Type        string
		SessionID   *string
		ReportedBy  string
		Description string
		OfflineSync bool
	}{
		LocationID:  req.LocationID,
		Type:        req.Type,
		SessionID:   req.SessionID,
		ReportedBy:  actorID,
		Description: req.Description,
	})
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "incident.created", inc.ID, &inc.LocationID, gin.H{
		"type":        inc.Type,
		"session_id":  inc.SessionID,
		"description": inc.Description,
	})

	response.Created(c, inc)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := store.ListIncidentsFilters{
		LocationID: c.Query("location_id"),
		Type:       c.Query("type"),
		State:      c.Query("state"),
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.DateTo = &t
		}
	}

	incidents, total, err := h.store.ListIncidents(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": incidents,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	inc, err := h.store.GetIncidentByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "incident")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, inc)
}

func (h *Handler) Resolve(c *gin.Context) {
	id := c.Param("id")

	var req ResolveIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	if req.AdjustmentAction != nil && *req.AdjustmentAction != "" {
		if req.ManagerPIN == "" {
			response.BadRequest(c, "MISSING_PIN", "manager PIN is required for adjustments")
			return
		}
		if err := h.validateManagerPIN(c.Request.Context(), actorID, req.ManagerPIN); err != nil {
			response.Forbidden(c, "invalid manager PIN")
			return
		}

		if req.AdjustmentEntityID == nil || *req.AdjustmentEntityID == "" {
			response.BadRequest(c, "MISSING_ENTITY", "adjustment_entity_id is required")
			return
		}

		switch *req.AdjustmentAction {
		case "VOID_TRANSACTION":
			tx, err := h.store.VoidTransaction(c.Request.Context(), *req.AdjustmentEntityID, actorID, "via incident resolution: "+req.ResolutionNotes)
			if err != nil {
				if err == errors.ErrNotFound {
					response.NotFound(c, "transaction")
					return
				}
				_ = c.Error(err)
				response.InternalServerError(c)
				return
			}
			_, _ = h.store.UpdateSessionToVoided(c.Request.Context(), tx.SessionID)
		case "REASSIGN_SESSION":
			tx, err := h.store.GetTransactionBySessionID(c.Request.Context(), *req.AdjustmentEntityID)
			if err != nil {
				if err == errors.ErrNotFound {
					response.NotFound(c, "session")
					return
				}
				_ = c.Error(err)
				response.InternalServerError(c)
				return
			}
			_, err = h.store.ReassignTransactionShift(c.Request.Context(), tx.ID, tx.OperatorID, tx.ShiftID)
			if err != nil {
				_ = c.Error(err)
				response.InternalServerError(c)
				return
			}
		}
	}

	inc, err := h.store.ResolveIncident(c.Request.Context(), id, actorID, req.ResolutionNotes, req.AdjustmentAction, req.AdjustmentEntityID)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "incident")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "incident.resolved", inc.ID, &inc.LocationID, gin.H{
		"resolution_notes":  req.ResolutionNotes,
		"adjustment_action": req.AdjustmentAction,
	})

	response.OK(c, inc)
}

func (h *Handler) AddNote(c *gin.Context) {
	id := c.Param("id")

	var req AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	note, err := h.store.CreateIncidentNote(c.Request.Context(), id, actorID, req.Note)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.Created(c, note)
}

func (h *Handler) ListNotes(c *gin.Context) {
	id := c.Param("id")
	notes, err := h.store.ListIncidentNotes(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, notes)
}

func (h *Handler) validateManagerPIN(ctx context.Context, userID, pin string) error {
	user, err := h.store.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.PINHash == nil || *user.PINHash == "" {
		return errors.ErrForbidden
	}
	if !authsvc.CheckPIN(pin, *user.PINHash) {
		return errors.ErrForbidden
	}
	return nil
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "incident",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}