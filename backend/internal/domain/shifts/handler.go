package shifts

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

type StartShiftRequest struct {
	LocationID string `json:"location_id" binding:"required,uuid"`
}

type EndShiftRequest struct {
	CashHandoverAmount float64 `json:"cash_handover_amount" binding:"gte=0"`
	DiscrepancyNotes   *string `json:"discrepancy_notes,omitempty"`
}

type ForceCloseShiftRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	shifts := r.Group("/shifts")
	shifts.Use(middleware.RequirePermission("shifts:view"))
	{
		shifts.GET("", h.List)
		shifts.GET("/me/open", h.GetMyOpen)
		shifts.GET("/:id", h.Get)
	}

	shiftsWithStart := r.Group("/shifts")
	shiftsWithStart.Use(middleware.RequirePermission("shifts:start"))
	{
		shiftsWithStart.POST("/start", h.Start)
	}

	shiftsWithEnd := r.Group("/shifts")
	shiftsWithEnd.Use(middleware.RequirePermission("shifts:end"))
	{
		shiftsWithEnd.POST("/:id/end", h.End)
	}

	shiftsWithForceClose := r.Group("/shifts")
	shiftsWithForceClose.Use(middleware.RequirePermission("shifts:force_close"))
	{
		shiftsWithForceClose.POST("/:id/force-close", h.ForceClose)
	}
}

func (h *Handler) GetMyOpen(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	shift, err := h.store.GetOpenShiftForOperator(c.Request.Context(), operatorID)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "no open shift")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, shift)
}

func (h *Handler) Start(c *gin.Context) {
	var req StartShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	ctx := c.Request.Context()

	// Auto-close any existing open shift for the operator.
	if existing, err := h.store.GetOpenShiftForOperator(ctx, operatorID); err == nil {
		_, _ = h.store.CloseShift(ctx, existing.ID, store.EndShiftInput{
			CashHandoverAmount: 0,
			DiscrepancyNotes:   strPtr("auto-closed because operator started a new shift"),
		})
		h.logAudit(c, "shift.auto_closed", existing.ID, &existing.LocationID, gin.H{
			"reason": "started_new_shift",
		})
	}

	shift, err := h.store.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: req.LocationID,
	})
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift.started", shift.ID, &req.LocationID, nil)
	response.Created(c, shift)
}

func (h *Handler) End(c *gin.Context) {
	id := c.Param("id")

	var req EndShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	shift, err := h.store.GetShiftByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	if shift.Status != "OPEN" {
		response.BadRequest(c, "INVALID_STATE", "shift is not open")
		return
	}

	expectedCash, err := h.store.SumCashByShift(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	// Persist expected cash before closing.
	_, err = h.store.UpdateShiftExpectedCash(c.Request.Context(), id, expectedCash)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	shift, err = h.store.CloseShift(c.Request.Context(), id, store.EndShiftInput{
		CashHandoverAmount: req.CashHandoverAmount,
		DiscrepancyNotes:   req.DiscrepancyNotes,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift.ended", shift.ID, &shift.LocationID, gin.H{
		"expected_cash":         expectedCash,
		"cash_handover_amount":  req.CashHandoverAmount,
		"discrepancy":           shift.Discrepancy,
	})

	response.OK(c, shift)
}

func (h *Handler) ForceClose(c *gin.Context) {
	id := c.Param("id")

	var req ForceCloseShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	shift, err := h.store.GetShiftByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	if shift.Status != "OPEN" {
		response.BadRequest(c, "INVALID_STATE", "shift is not open")
		return
	}

	actorID := middleware.GetUserID(c)
	shift, err = h.store.ForceCloseShift(c.Request.Context(), id, store.ForceCloseShiftInput{
		ForceClosedBy:     actorID,
		ForceClosedReason: req.Reason,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift.force_closed", shift.ID, &shift.LocationID, gin.H{
		"reason":          req.Reason,
		"force_closed_by": actorID,
	})

	response.OK(c, shift)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := map[string]interface{}{}
	if v := c.Query("location_id"); v != "" {
		filters["location_id"] = v
	}
	if v := c.Query("operator_id"); v != "" {
		filters["operator_id"] = v
	}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters["date_from"] = t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters["date_to"] = t
		}
	}

	shifts, total, err := h.store.ListShifts(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": shifts,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.store.GetShiftByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	include := c.Query("include")
	if include == "transactions" {
		transactions, total, err := h.store.ListTransactions(c.Request.Context(), store.ListTransactionsFilters{
			ShiftID: shift.ID,
		}, 100, 0)
		if err != nil {
			_ = c.Error(err)
			response.InternalServerError(c)
			return
		}

		expectedCash, err := h.store.SumCashByShift(c.Request.Context(), shift.ID)
		if err != nil {
			_ = c.Error(err)
			response.InternalServerError(c)
			return
		}

		response.OK(c, gin.H{
			"shift":        shift,
			"transactions": transactions,
			"summary": gin.H{
				"transaction_count": total,
				"expected_cash":     expectedCash,
			},
		})
		return
	}

	response.OK(c, shift)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "shift",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}

func strPtr(s string) *string {
	return &s
}
