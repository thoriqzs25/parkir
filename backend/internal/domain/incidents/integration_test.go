package incidents_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
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

	// Create shift config and get shift instance
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
	
	shift, err := s.GetOrCreateShift(ctx, locationID, 1, time.Now())
	if err != nil {
		t.Fatalf("get or create shift: %v", err)
	}

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
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
		ShiftID:              shift.ID,
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

func TestReassignSession(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := newSeedLocation(ctx, t, s)
	operator1 := newSeedUser(ctx, t, s, "operator")
	operator2 := newSeedUser(ctx, t, s, "operator")

	// Create shift configs
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

	shift1, err := s.GetOrCreateShift(ctx, locationID, 1, time.Now())
	if err != nil {
		t.Fatalf("get or create shift 1: %v", err)
	}

	// For operator2, use same shift (static shift model)
	shift2 := shift1

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operator1,
		ShiftID:     shift1.ID,
		Plate:       "B5678ABC",
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.OperatorID != operator1 {
		t.Fatalf("expected operator1 as operator")
	}

	// Reassign to operator2
	reassignedSession, err := s.ReassignSession(ctx, session.ID, operator2, shift2.ID)
	if err != nil {
		t.Fatalf("reassign session: %v", err)
	}
	if reassignedSession.OperatorID != operator2 {
		t.Fatalf("expected operator2 as operator, got %s", reassignedSession.OperatorID)
	}
}

func TestAlertLifecycle(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	// Create alert
	alert, err := s.CreateAlert(ctx, struct {
		Code       string
		LocationID *string
		EntityType *string
		EntityID   *string
		Metadata   map[string]interface{}
	}{
		Code:       "LONG_SESSION",
		LocationID: nil,
		EntityType: strPtr("session"),
		EntityID:   strPtr("00000000-0000-0000-0000-000000000001"),
		Metadata:   map[string]interface{}{"plate": "B9999XX"},
	})
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}
	if alert.State != "TRIGGERED" {
		t.Fatalf("expected TRIGGERED, got %s", alert.State)
	}

	// Acknowledge
	acknowledgedBy := newSeedUser(ctx, t, s, "user")
	ackd, err := s.AcknowledgeAlert(ctx, alert.ID, acknowledgedBy)
	if err != nil {
		t.Fatalf("acknowledge alert: %v", err)
	}
	if ackd.State != "ACKNOWLEDGED" {
		t.Fatalf("expected ACKNOWLEDGED, got %s", ackd.State)
	}

	// Resolve
	resolvedBy := newSeedUser(ctx, t, s, "user")
	resolved, err := s.ResolveAlert(ctx, alert.ID, resolvedBy, "Issue resolved")
	if err != nil {
		t.Fatalf("resolve alert: %v", err)
	}
	if resolved.State != "RESOLVED" {
		t.Fatalf("expected RESOLVED, got %s", resolved.State)
	}

	// Count triggered
	count, err := s.CountTriggeredAlerts(ctx)
	if err != nil {
		t.Fatalf("count triggered alerts: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 triggered alerts, got %d", count)
	}
}

func TestAuditLogQuery(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	// Create some audit logs
	actorID := newSeedUser(ctx, t, s, "user")
	entityID := "00000000-0000-0000-0000-000000000099"
	locationID := newSeedLocation(ctx, t, s)

	for i := 0; i < 5; i++ {
		err := s.CreateAuditLog(ctx, store.AuditLogEntry{
			Action:     "test.action",
			ActorID:    &actorID,
			ActorRole:  strPtr("tester"),
			EntityType: "test",
			EntityID:   entityID,
			LocationID: &locationID,
			IPAddress:  strPtr("127.0.0.1"),
			Metadata:   map[string]interface{}{"index": i},
		})
		if err != nil {
			t.Fatalf("create audit log: %v", err)
		}
	}

	// Query with filter
	filters := store.ListAuditLogsFilters{
		LocationID: locationID,
		Action:     "test.action",
	}
	logs, total, err := s.ListAuditLogs(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected 5 audit logs, got %d", total)
	}
	if len(logs) != 5 {
		t.Fatalf("expected 5 audit logs in result, got %d", len(logs))
	}

	// Test all filter
	allLogs, err := s.ListAuditLogsAll(ctx, filters)
	if err != nil {
		t.Fatalf("list all audit logs: %v", err)
	}
	if len(allLogs) != 5 {
		t.Fatalf("expected 5 audit logs from all query, got %d", len(allLogs))
	}
}

func TestPINValidation(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	// Create manager with PIN
	roleID := newEnsureRole(ctx, t, s, "manager", []string{"*"})
	pinHash, err := authsvc.HashPIN("123456")
	if err != nil {
		t.Fatalf("hash pin: %v", err)
	}

	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         "Manager With PIN",
		Email:        fmt.Sprintf("manager-pin-%d@test.local", time.Now().UnixNano()),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	err = s.UpdatePIN(ctx, user.ID, pinHash)
	if err != nil {
		t.Fatalf("update PIN: %v", err)
	}

	// Verify correct PIN
	dbUser, err := s.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if dbUser.PINHash == nil || !authsvc.CheckPIN("123456", *dbUser.PINHash) {
		t.Fatalf("expected PIN 123456 to be valid")
	}

	// Verify wrong PIN
	if authsvc.CheckPIN("000000", *dbUser.PINHash) {
		t.Fatalf("expected PIN 000000 to be invalid")
	}
}

func TestHasAlertForEntity(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	// No alert yet
	exists, err := s.HasAlertForEntity(ctx, "LONG_SESSION", "session", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("has alert for entity: %v", err)
	}
	if exists {
		t.Fatalf("expected no existing alert")
	}

	// Create alert
	_, err = s.CreateAlert(ctx, struct {
		Code       string
		LocationID *string
		EntityType *string
		EntityID   *string
		Metadata   map[string]interface{}
	}{
		Code:       "LONG_SESSION",
		EntityType: strPtr("session"),
		EntityID:   strPtr("00000000-0000-0000-0000-000000000001"),
	})
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}

	// Now exists
	exists, err = s.HasAlertForEntity(ctx, "LONG_SESSION", "session", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("has alert for entity: %v", err)
	}
	if !exists {
		t.Fatalf("expected alert to exist")
	}

	// Wrong entity ID should not exist
	exists, err = s.HasAlertForEntity(ctx, "LONG_SESSION", "session", "00000000-0000-0000-0000-000000000002")
	if err != nil {
		t.Fatalf("has alert for entity: %v", err)
	}
	if exists {
		t.Fatalf("expected no alert for different entity")
	}
}

// Helpers

func newSeedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	loc, err := s.CreateLocation(ctx, store.CreateLocationInput{
		Name: fmt.Sprintf("Test Location %d", time.Now().UnixNano()),
		Code: fmt.Sprintf("TST%d", time.Now().UnixNano()%10000),
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc.ID
}

func newSeedUser(ctx context.Context, t *testing.T, s *store.Store, roleName string) string {
	t.Helper()
	roleID := newEnsureRole(ctx, t, s, roleName, []string{"*"})
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         fmt.Sprintf("User %d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("user-%d@test.local", time.Now().UnixNano()),
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
	roleID := newEnsureRole(ctx, t, s, "manager", []string{"*"})
	pinHash, err := authsvc.HashPIN("123456")
	if err != nil {
		t.Fatalf("hash pin: %v", err)
	}

	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         fmt.Sprintf("Manager %d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("manager-%d@test.local", time.Now().UnixNano()),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}

	if err := s.UpdatePIN(ctx, user.ID, pinHash); err != nil {
		t.Fatalf("set PIN: %v", err)
	}

	return user.ID
}

func newEnsureRole(ctx context.Context, t *testing.T, s *store.Store, name string, permissions []string) string {
	t.Helper()
	var id string
	err := s.Pool().QueryRow(ctx, `
		INSERT INTO roles (name, permissions)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		RETURNING id
	`, name, permissions).Scan(&id)
	if err != nil {
		t.Fatalf("ensure role %s: %v", name, err)
	}
	return id
}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
