package incidents_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/store"
	"github.com/thoriqzs/PARKIR/backend/internal/testutil"
)

func TestIncidentLifecycle(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := newSeedLocation(ctx, t, s)
	operatorID := newSeedUser(ctx, t, s, "operator")
	managerID := newSeedManager(ctx, t, s, locationID)

	// Create incident
	inc, err := s.CreateIncident(ctx, struct {
		LocationID  string
		Type        string
		SessionID   *string
		ReportedBy  string
		Description string
		OfflineSync bool
	}{
		LocationID:  locationID,
		Type:        "OPERATOR_ERROR",
		ReportedBy:  operatorID,
		Description: "Test incident: operator entered wrong plate",
	})
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	if inc.State != "OPEN" {
		t.Fatalf("expected OPEN state, got %s", inc.State)
	}

	// Add note
	note, err := s.CreateIncidentNote(ctx, inc.ID, managerID, "Investigated the issue")
	if err != nil {
		t.Fatalf("create incident note: %v", err)
	}
	if note.IncidentID != inc.ID {
		t.Fatalf("note incident_id mismatch")
	}

	// List notes
	notes, err := s.ListIncidentNotes(ctx, inc.ID)
	if err != nil {
		t.Fatalf("list incident notes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	// Resolve without adjustment
	resolved, err := s.ResolveIncident(ctx, inc.ID, managerID, "Resolved - retrained operator", nil, nil)
	if err != nil {
		t.Fatalf("resolve incident: %v", err)
	}
	if resolved.State != "RESOLVED" {
		t.Fatalf("expected RESOLVED, got %s", resolved.State)
	}

	// List with filter
	filters := store.ListIncidentsFilters{
		LocationID: locationID,
		State:      "RESOLVED",
	}
	incidents, total, err := s.ListIncidents(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("list incidents: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 resolved incident, got %d", total)
	}
	if len(incidents) != 1 {
		t.Fatalf("expected 1 incident in result, got %d", len(incidents))
	}
}

func TestVoidTransactionAdjustment(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := newSeedLocation(ctx, t, s)
	operatorID := newSeedUser(ctx, t, s, "operator")
	managerID := newSeedManager(ctx, t, s, locationID)

	// Create shift config
	_, err := s.CreateLocationShiftConfig(ctx, store.CreateLocationShiftConfigInput{
		LocationID:  locationID,
		ShiftCode:   "08-16",
		ShiftNumber: 1,
		StartTime:   "08:00:00",
		EndTime:     "16:00:00",
	})
	if err != nil {
		t.Fatalf("create shift config: %v", err)
	}

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftNumber: 1,
		Plate:       "B1234XYZ",
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Set to pending payment
	fee := 5000.0
	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: time.Now(),
		FeeAmount:  &fee,
	})
	if err != nil {
		t.Fatalf("checkout session: %v", err)
	}

	// Create a transaction
	receiptNumber := fmt.Sprintf("TST-%s-%05d", time.Now().Format("20060102"), 1)
	tx, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:            session.ID,
		LocationID:           locationID,
		ShiftNumber:          1,
		OperatorID:           operatorID,
		VehicleType:          "CAR",
		Plate:                "B1234XYZ",
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           time.Now(),
		DurationHours:        1,
		RateFirstHour:        5000,
		RateSubsequentHourly: 3000,
		RateDaily:            50000,
		FeeAmount:            5000,
		PaymentMethod:        "CASH",
		AmountTendered:       &fee,
		ChangeAmount:         float64Ptr(0),
		ReceiptNumber:        receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if tx.Voided {
		t.Fatalf("expected non-voided transaction")
	}

	// Void transaction via adjustment
	voidedTx, err := s.VoidTransaction(ctx, tx.ID, managerID, "Test void via adjustment")
	if err != nil {
		t.Fatalf("void transaction: %v", err)
	}
	if !voidedTx.Voided {
		t.Fatalf("expected voided transaction")
	}

	// Also void the session
	_, err = s.UpdateSessionToVoided(ctx, session.ID)
	if err != nil {
		t.Fatalf("void session: %v", err)
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func newSeedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	loc, err := s.CreateLocation(ctx, store.CreateLocationInput{
		Name:    "Test Location",
		Code:    fmt.Sprintf("LOC-%d", time.Now().UnixNano()),
		Address: "Test Address",
		City:    "Test City",
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc.ID
}

func newSeedUser(ctx context.Context, t *testing.T, s *store.Store, roleName string) string {
	t.Helper()
	roleID := ensureRole(ctx, t, s, roleName)
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         fmt.Sprintf("Test %s", roleName),
		Email:        fmt.Sprintf("%s-%d@test.local", roleName, time.Now().UnixNano()),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func newSeedManager(ctx context.Context, t *testing.T, s *store.Store, locationID string) string {
	t.Helper()
	return newSeedUser(ctx, t, s, "manager")
}

func ensureRole(ctx context.Context, t *testing.T, s *store.Store, roleName string) string {
	t.Helper()
	permissions := []string{}
	switch roleName {
	case "operator":
		permissions = []string{"sessions:*", "payments:*"}
	case "manager":
		permissions = []string{"sessions:*", "payments:*", "adjustments:*", "incidents:*"}
	}

	var id string
	err := s.Pool().QueryRow(ctx, `
		INSERT INTO roles (name, permissions)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		RETURNING id
	`, roleName, permissions).Scan(&id)
	if err != nil {
		t.Fatalf("ensure role: %v", err)
	}
	return id
}
