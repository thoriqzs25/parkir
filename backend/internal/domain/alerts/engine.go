package alerts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/notifier"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Engine struct {
	store    *store.Store
	notifier *notifier.Notifier
}

func NewEngine(store *store.Store, notifier *notifier.Notifier) *Engine {
	return &Engine{
		store:    store,
		notifier: notifier,
	}
}

func (e *Engine) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.runChecks(ctx)

	for {
		select {
		case <-ticker.C:
			e.runChecks(ctx)
		case <-ctx.Done():
			log.Println("alert engine stopped")
			return
		}
	}
}

func (e *Engine) runChecks(ctx context.Context) {
	e.checkLongSessions(ctx)
	e.checkUnpaidExits(ctx)
}

func (e *Engine) checkLongSessions(ctx context.Context) {
	rows, err := e.store.Pool().Query(ctx, `
		SELECT s.id, s.plate, s.check_in_at, s.location_id, l.name as location_name
		FROM sessions s
		JOIN locations l ON l.id = s.location_id
		WHERE s.state = 'ACTIVE'
		  AND s.check_in_at < now() - interval '24 hours'
	`)
	if err != nil {
		log.Printf("alert engine: query long sessions: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID, plate, locationID, locationName string
		var checkInAt time.Time
		if err := rows.Scan(&sessionID, &plate, &checkInAt, &locationID, &locationName); err != nil {
			log.Printf("alert engine: scan long session: %v", err)
			continue
		}

		exists, err := e.store.HasAlertForEntity(ctx, "LONG_SESSION", "session", sessionID)
		if err != nil || exists {
			continue
		}

		duration := time.Since(checkInAt).Hours()
		description := fmt.Sprintf("Session %s (%s) has been active for %.0f hours at %s", plate, sessionID, duration, locationName)

		alert, err := e.store.CreateAlert(ctx, struct {
			Code       string
			LocationID *string
			EntityType *string
			EntityID   *string
			Metadata   map[string]interface{}
		}{
			Code:       "LONG_SESSION",
			LocationID: &locationID,
			EntityType: strPtr("session"),
			EntityID:   &sessionID,
			Metadata: map[string]interface{}{
				"plate":       plate,
				"check_in_at": checkInAt,
				"duration_hours": duration,
			},
		})
		if err != nil {
			log.Printf("alert engine: create long session alert: %v", err)
			continue
		}

		log.Printf("alert engine: LONG_SESSION alert created: %s", alert.ID)

		if err := e.notifier.SendAlertEmail("LONG_SESSION", locationName, description); err != nil {
			log.Printf("alert engine: send email for LONG_SESSION: %v", err)
		}
	}
}

func (e *Engine) checkUnpaidExits(ctx context.Context) {
	rows, err := e.store.Pool().Query(ctx, `
		SELECT s.id, s.plate, s.check_out_at, s.location_id, l.name as location_name
		FROM sessions s
		JOIN locations l ON l.id = s.location_id
		WHERE s.state = 'PENDING_PAYMENT'
		  AND s.check_out_at < now() - interval '30 minutes'
	`)
	if err != nil {
		log.Printf("alert engine: query unpaid exits: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID, plate, locationID, locationName string
		var checkOutAt time.Time
		if err := rows.Scan(&sessionID, &plate, &checkOutAt, &locationID, &locationName); err != nil {
			log.Printf("alert engine: scan unpaid exit: %v", err)
			continue
		}

		exists, err := e.store.HasAlertForEntity(ctx, "UNPAID_EXIT", "session", sessionID)
		if err != nil || exists {
			continue
		}

		description := fmt.Sprintf("Session %s (%s) is in PENDING_PAYMENT since %s at %s", plate, sessionID, checkOutAt.Format(time.RFC3339), locationName)

		alert, err := e.store.CreateAlert(ctx, struct {
			Code       string
			LocationID *string
			EntityType *string
			EntityID   *string
			Metadata   map[string]interface{}
		}{
			Code:       "UNPAID_EXIT",
			LocationID: &locationID,
			EntityType: strPtr("session"),
			EntityID:   &sessionID,
			Metadata: map[string]interface{}{
				"plate":       plate,
				"check_out_at": checkOutAt,
			},
		})
		if err != nil {
			log.Printf("alert engine: create unpaid exit alert: %v", err)
			continue
		}

		log.Printf("alert engine: UNPAID_EXIT alert created: %s", alert.ID)

		if err := e.notifier.SendAlertEmail("UNPAID_EXIT", locationName, description); err != nil {
			log.Printf("alert engine: send email for UNPAID_EXIT: %v", err)
		}
	}
}

func strPtr(s string) *string {
	return &s
}