package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

// Handler exposes the sync API used by the desktop app to push offline records.
type Handler struct {
	store *store.Store
}

// NewHandler creates a sync handler.
func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

// SyncItem represents one operation in a sync batch.
type SyncItem struct {
	Type            string                   `json:"type" binding:"required,oneof=check_in check_out payment"`
	SessionID       string                   `json:"session_id,omitempty"`
	TransactionID   string                   `json:"transaction_id,omitempty"`
	LocationID      string                   `json:"location_id,omitempty"`
	Session         *OfflineSessionData      `json:"session,omitempty"`
	CheckOutAt      *time.Time               `json:"check_out_at,omitempty"`
	FeeAmount       *float64                 `json:"fee_amount,omitempty"`
	RateSnapshot    map[string]interface{}   `json:"rate_snapshot,omitempty"`
	ShiftID         string                   `json:"shift_id,omitempty"`
	OperatorID      string                   `json:"operator_id,omitempty"`
	DurationHours   int                      `json:"duration_hours,omitempty"`
	RateFirstHour   float64                  `json:"rate_first_hour,omitempty"`
	RateSubsequent  float64                  `json:"rate_subsequent_hourly,omitempty"`
	RateDaily       float64                  `json:"rate_daily,omitempty"`
	PaymentMethod   string                   `json:"payment_method,omitempty"`
	AmountTendered  *float64                 `json:"amount_tendered,omitempty"`
	ChangeAmount      *float64                 `json:"change_amount,omitempty"`
	PaymentReference  *string                  `json:"payment_reference,omitempty"`
}

// OfflineSessionData carries the full session payload for a check_in operation.
type OfflineSessionData struct {
	ID          string    `json:"id" binding:"required,uuid"`
	LocationID  string    `json:"location_id" binding:"required,uuid"`
	OperatorID  string    `json:"operator_id" binding:"required,uuid"`
	ShiftID     string    `json:"shift_id" binding:"required,uuid"`
	Plate       string    `json:"plate" binding:"required"`
	CityCode    string    `json:"city_code"`
	VehicleType string    `json:"vehicle_type" binding:"required"`
	CheckInAt   time.Time `json:"check_in_at" binding:"required"`
}

// BatchSyncRequest is the top-level payload sent from the desktop app.
type BatchSyncRequest struct {
	Items []SyncItem `json:"items" binding:"required,dive"`
}

// SyncResult describes the outcome of a single sync item.
type SyncResult struct {
	Type            string  `json:"type"`
	SessionID       string  `json:"session_id,omitempty"`
	TransactionID   string  `json:"transaction_id,omitempty"`
	ReceiptNumber   string  `json:"receipt_number,omitempty"`
	SyncConflict    bool    `json:"sync_conflict"`
	Error           string  `json:"error,omitempty"`
}

// BatchSyncResponse wraps all results for a batch.
type BatchSyncResponse struct {
	Results []SyncResult `json:"results"`
}

// RegisterRoutes registers sync endpoints on the given router group.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	sync := r.Group("/sync")
	sync.Use(middleware.RequirePermission("sessions:create"))
	{
		sync.POST("/batch", h.BatchSync)
	}

	conflicts := r.Group("/sync/conflicts")
	conflicts.Use(middleware.RequirePermission("sessions:view"))
	{
		conflicts.GET("", h.ListConflicts)
		conflicts.POST("/:id/resolve", middleware.RequirePermission("sessions:void"), h.ResolveConflict)
	}
}

// BatchSync processes a batch of offline operations submitted by a desktop client.
func (h *Handler) BatchSync(c *gin.Context) {
	var req BatchSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	results := make([]SyncResult, 0, len(req.Items))
	for _, item := range req.Items {
		result := h.processItem(c.Request.Context(), item)
		results = append(results, result)
	}

	response.OK(c, BatchSyncResponse{Results: results})
}

func (h *Handler) processItem(ctx context.Context, item SyncItem) SyncResult {
	result := SyncResult{Type: item.Type}

	switch item.Type {
	case "check_in":
		if item.Session == nil {
			result.Error = "missing session data"
			return result
		}
		data := item.Session

		exists, err := h.store.ValidateVehicleTypeExists(ctx, data.VehicleType)
		if err != nil || !exists {
			result.Error = "unknown vehicle type"
			return result
		}

		session, err := h.store.CreateOfflineSession(ctx, store.CreateOfflineSessionInput{
			ID:          data.ID,
			LocationID:  data.LocationID,
			OperatorID:  data.OperatorID,
			ShiftID:     data.ShiftID,
			Plate:       normalizePlate(data.Plate),
			CityCode:    normalizeCode(data.CityCode),
			VehicleType: data.VehicleType,
			CheckInAt:   data.CheckInAt,
		})
		if err != nil {
			result.Error = err.Error()
			return result
		}
		result.SessionID = session.ID
		result.SyncConflict = session.SyncConflict

	case "check_out":
		if item.SessionID == "" || item.CheckOutAt == nil || item.FeeAmount == nil {
			result.Error = "missing checkout fields"
			return result
		}
		result.SessionID = item.SessionID
		_, err := h.store.UpdateSessionToPendingPayment(ctx, item.SessionID, store.CheckOutSessionInput{
			CheckOutAt:   *item.CheckOutAt,
			FeeAmount:    item.FeeAmount,
			RateSnapshot: item.RateSnapshot,
		})
		if err != nil {
			result.Error = err.Error()
		}

	case "payment":
		if item.SessionID == "" || item.TransactionID == "" || item.ShiftID == "" || item.OperatorID == "" || item.PaymentMethod == "" || item.LocationID == "" || item.FeeAmount == nil {
			result.Error = "missing payment fields"
			return result
		}
		receiptNumber, err := h.store.GenerateReceiptNumber(ctx, item.LocationID, time.Now().UTC())
		if err != nil {
			result.Error = fmt.Sprintf("generate receipt: %v", err)
			return result
		}

		tx, err := h.store.CreateOfflineTransaction(ctx, store.CreateOfflineTransactionInput{
			ID:                   item.TransactionID,
			SessionID:            item.SessionID,
			ShiftID:              item.ShiftID,
			OperatorID:           item.OperatorID,
			DurationHours:        item.DurationHours,
			RateFirstHour:        item.RateFirstHour,
			RateSubsequentHourly: item.RateSubsequent,
			RateDaily:            item.RateDaily,
			FeeAmount:            *item.FeeAmount,
			PaymentMethod:        item.PaymentMethod,
			AmountTendered:       item.AmountTendered,
			ChangeAmount:         item.ChangeAmount,
			PaymentReference:     item.PaymentReference,
			ReceiptNumber:        receiptNumber,
		})
		if err != nil {
			result.Error = err.Error()
			return result
		}
		result.SessionID = tx.SessionID
		result.TransactionID = tx.ID
		result.ReceiptNumber = tx.ReceiptNumber
	}

	return result
}

func normalizePlate(plate string) string {
	return normalizeCode(plate)
}

func normalizeCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return "UNKNOWN"
	}
	return code
}

// ListConflicts returns sessions flagged with sync conflicts for manager review.
func (h *Handler) ListConflicts(c *gin.Context) {
	limit := parseInt(c.DefaultQuery("limit", "20"), 20)
	offset := parseInt(c.DefaultQuery("offset", "0"), 0)

	filters := store.ListSyncConflictsFilters{}
	if loc := c.Query("location_id"); loc != "" {
		filters.LocationID = loc
	}

	sessions, total, err := h.store.ListSyncConflicts(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": sessions,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

// ResolveConflictRequest carries a manager's resolution for a conflict.
type ResolveConflictRequest struct {
	Action     string `json:"action" binding:"required,oneof=VOID_OFFLINE IGNORE"`
	VoidReason string `json:"void_reason"`
}

// ResolveConflict applies a manager's resolution to a conflicting session.
func (h *Handler) ResolveConflict(c *gin.Context) {
	id := c.Param("id")

	var req ResolveConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	session, err := h.store.ResolveSyncConflict(c.Request.Context(), store.ResolveSyncConflictInput{
		SessionID:  id,
		Action:     store.ResolveSyncConflictAction(req.Action),
		VoidReason: req.VoidReason,
		ResolvedBy: actorID,
	})
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			response.NotFound(c, "session")
		case errors.ErrInvalidState:
			response.BadRequest(c, "INVALID_STATE", "session is not a sync conflict")
		case errors.ErrInvalidInput:
			response.BadRequest(c, "INVALID_ACTION", "unknown resolution action")
		default:
			_ = c.Error(err)
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, session)
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return defaultValue
	}
	return n
}
