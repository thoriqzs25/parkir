package transactions

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

type CashPaymentRequest struct {
	SessionID      string  `json:"session_id" binding:"required,uuid"`
	AmountTendered float64 `json:"amount_tendered" binding:"required,gte=0"`
}

type DigitalPaymentRequest struct {
	SessionID        string  `json:"session_id" binding:"required,uuid"`
	PaymentReference *string `json:"payment_reference,omitempty"`
}

type VoidTransactionRequest struct {
	ManagerPIN  string `json:"manager_pin" binding:"required,len=6"`
	VoidReason  string `json:"void_reason" binding:"required"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	payments := r.Group("/payments")
	{
		payments.POST("/cash", middleware.RequirePermission("payments:collect_cash"), h.Cash)
		payments.POST("/digital", middleware.RequirePermission("payments:collect_digital"), h.Digital)
	}

	transactions := r.Group("/transactions")
	transactions.Use(middleware.RequirePermission("sessions:view"))
	{
		transactions.GET("", h.List)
		transactions.GET("/:id", h.Get)
	}

	transactionsWithVoid := r.Group("/transactions")
	transactionsWithVoid.Use(middleware.RequirePermission("payments:void"))
	{
		transactionsWithVoid.POST("/:id/void", h.Void)
	}
}

func (h *Handler) Cash(c *gin.Context) {
	var req CashPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	session, tx, err := h.recordPayment(c, req.SessionID, "CASH", &req.AmountTendered, nil)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	h.logAudit(c, "transaction.cash", tx.ID, &tx.LocationID, gin.H{
		"session_id":      session.ID,
		"amount_tendered": req.AmountTendered,
		"change":          tx.ChangeAmount,
		"receipt_number":  tx.ReceiptNumber,
	})

	response.Created(c, tx)
}

func (h *Handler) Digital(c *gin.Context) {
	var req DigitalPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	session, tx, err := h.recordPayment(c, req.SessionID, "DIGITAL", nil, req.PaymentReference)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	h.logAudit(c, "transaction.digital", tx.ID, &tx.LocationID, gin.H{
		"session_id":       session.ID,
		"payment_reference": tx.PaymentReference,
		"receipt_number":   tx.ReceiptNumber,
	})

	response.Created(c, tx)
}

func (h *Handler) recordPayment(c *gin.Context, sessionID, method string, amountTendered *float64, paymentReference *string) (*store.Session, *store.Transaction, error) {
	ctx := c.Request.Context()

	session, err := h.store.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if session.State != "PENDING_PAYMENT" {
		return nil, nil, errors.ErrInvalidState
	}

	operatorID := middleware.GetUserID(c)

	shift, err := h.store.GetOpenShiftForOperator(ctx, operatorID)
	if err != nil {
		return nil, nil, err
	}
	if shift.LocationID != session.LocationID {
		return nil, nil, errors.ErrShiftLocationMismatch
	}

	fee := *session.FeeAmount
	var change *float64
	if method == "CASH" && amountTendered != nil {
		if *amountTendered < fee {
			return nil, nil, errors.ErrInsufficientPayment
		}
		ch := *amountTendered - fee
		change = &ch
	}

	receiptNumber, err := h.store.GenerateReceiptNumber(ctx, session.LocationID, time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}

	durationHours := 1
	if session.CheckOutAt != nil {
		durationHours = calculateDurationHours(session.CheckInAt, *session.CheckOutAt)
	}

	rateSnapshot := session.RateSnapshot
	firstHour := getRateValue(rateSnapshot, "first_hour_rate")
	subsequentHour := getRateValue(rateSnapshot, "subsequent_hourly_rate")
	daily := getRateValue(rateSnapshot, "daily_flat_rate")

	if firstHour == 0 && subsequentHour == 0 && daily == 0 {
		firstHour = fee
		subsequentHour = fee
		daily = fee
	}

	tx, err := h.store.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:            session.ID,
		LocationID:           session.LocationID,
		ShiftID:              shift.ID,
		OperatorID:           operatorID,
		VehicleType:          session.VehicleType,
		Plate:                session.Plate,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           *session.CheckOutAt,
		DurationHours:        durationHours,
		RateFirstHour:        firstHour,
		RateSubsequentHourly: subsequentHour,
		RateDaily:            daily,
		FeeAmount:            fee,
		PaymentMethod:        method,
		AmountTendered:       amountTendered,
		ChangeAmount:         change,
		PaymentReference:     paymentReference,
		ReceiptNumber:        receiptNumber,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = h.store.UpdateSessionToClosed(ctx, session.ID)
	if err != nil {
		return nil, nil, err
	}

	return session, tx, nil
}

func (h *Handler) handlePaymentError(c *gin.Context, err error) {
	_ = c.Error(err)
	switch err {
	case errors.ErrNotFound:
		response.NotFound(c, "session")
	case errors.ErrInvalidState:
		response.BadRequest(c, "INVALID_STATE", "session is not pending payment")
	case errors.ErrShiftLocationMismatch:
		response.BadRequest(c, "SHIFT_LOCATION_MISMATCH", "operator shift is for a different location")
	case errors.ErrInsufficientPayment:
		response.BadRequest(c, "INSUFFICIENT_PAYMENT", "amount tendered is less than fee")
	default:
		response.InternalServerError(c)
	}
}

func (h *Handler) Void(c *gin.Context) {
	id := c.Param("id")

	var req VoidTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	// Validate manager PIN. Managers/admins/owners have PINs.
	if err := h.validateManagerPIN(c.Request.Context(), actorID, req.ManagerPIN); err != nil {
		response.Forbidden(c, "invalid manager PIN")
		return
	}

	tx, err := h.store.VoidTransaction(c.Request.Context(), id, actorID, req.VoidReason)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "transaction")
			return
		}
		response.InternalServerError(c)
		return
	}

	_, err = h.store.UpdateSessionToVoided(c.Request.Context(), tx.SessionID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "transaction.voided", tx.ID, &tx.LocationID, gin.H{
		"session_id":    tx.SessionID,
		"void_reason":   req.VoidReason,
		"voided_by":     actorID,
		"receipt_number": tx.ReceiptNumber,
	})

	response.OK(c, tx)
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

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := store.ListTransactionsFilters{
		LocationID: c.Query("location_id"),
		ShiftID:    c.Query("shift_id"),
	}
	if v := c.Query("voided"); v != "" {
		b := v == "true"
		filters.Voided = &b
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

	transactions, total, err := h.store.ListTransactions(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": transactions,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	tx, err := h.store.GetTransactionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "transaction")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, tx)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "transaction",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}

func calculateDurationHours(checkInAt, checkOutAt time.Time) int {
	diff := checkOutAt.Sub(checkInAt)
	hours := int(diff.Hours())
	if diff.Hours() > float64(hours) {
		hours++
	}
	if hours < 1 {
		hours = 1
	}
	return hours
}

func getRateValue(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
