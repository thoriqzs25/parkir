package adjustments

import (
	"context"

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

type VoidTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required,uuid"`
	Reason        string `json:"reason" binding:"required"`
	ManagerPIN    string `json:"manager_pin" binding:"required,len=6"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	adj := r.Group("/adjustments")
	adj.Use(middleware.RequirePermission("adjustments:void_transaction"))
	{
		adj.POST("/void-transaction", h.VoidTransaction)
	}
}

func (h *Handler) VoidTransaction(c *gin.Context) {
	var req VoidTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	if err := h.validateManagerPIN(c.Request.Context(), actorID, req.ManagerPIN); err != nil {
		response.Forbidden(c, "invalid manager PIN")
		return
	}

	tx, err := h.store.VoidTransaction(c.Request.Context(), req.TransactionID, actorID, req.Reason)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "transaction")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	_, err = h.store.UpdateSessionToVoided(c.Request.Context(), tx.SessionID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     "adjustment.void_transaction",
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "transaction",
		EntityID:   tx.ID,
		LocationID: &tx.LocationID,
		IPAddress:  &ip,
		Metadata: map[string]interface{}{
			"session_id":     tx.SessionID,
			"reason":         req.Reason,
			"receipt_number": tx.ReceiptNumber,
		},
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
