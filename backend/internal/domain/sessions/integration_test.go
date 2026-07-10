package sessions_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/fees"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
	"github.com/thoriqzs/PARKIR/backend/internal/testutil"
)

func TestFullParkingFlow(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	// Seed a location, operator, rate, and shift config.
	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	seedShiftConfig(ctx, t, s, locationID)
	
	// Get or create shift instance
	shift, err := s.GetOrCreateShift(ctx, locationID, 1, time.Now())
	if err != nil {
		t.Fatalf("get or create shift: %v", err)
	}

	// Check in.
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
	if session.State != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", session.State)
	}

	// Check out after 3 hours.
	checkOutAt := session.CheckInAt.Add(3 * time.Hour)
	expectedFee, _ := fees.Calculate(fees.CalculationInput{
		FirstHourRate:        5000,
		SubsequentHourlyRate: 3000,
		DailyFlatRate:        50000,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           checkOutAt,
	})
	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &expectedFee,
	})
	if err != nil {
		t.Fatalf("checkout session: %v", err)
	}
	if session.State != "PENDING_PAYMENT" {
		t.Fatalf("expected PENDING_PAYMENT, got %s", session.State)
	}
	if session.FeeAmount == nil || *session.FeeAmount != expectedFee {
		t.Fatalf("expected fee %v, got %v", expectedFee, session.FeeAmount)
	}

	// Pay and close.
	receiptNumber := generateReceiptNumber()
	transaction, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:            session.ID,
		LocationID:           locationID,
		ShiftID:              shift.ID,
		OperatorID:           operatorID,
		VehicleType:          "CAR",
		Plate:                session.Plate,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           checkOutAt,
		DurationHours:        3,
		RateFirstHour:        5000,
		RateSubsequentHourly: 3000,
		RateDaily:            50000,
		FeeAmount:            expectedFee,
		PaymentMethod:        "CASH",
		AmountTendered:       &expectedFee,
		ReceiptNumber:        receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	_, err = s.UpdateSessionToClosed(ctx, session.ID)
	if err != nil {
		t.Fatalf("close session: %v", err)
	}

	// Void the transaction and verify session becomes VOIDED.
	_, err = s.VoidTransaction(ctx, transaction.ID, operatorID, "customer complaint")
	if err != nil {
		t.Fatalf("void transaction: %v", err)
	}
	voidedSession, err := s.UpdateSessionToVoided(ctx, session.ID)
	if err != nil {
		t.Fatalf("void session: %v", err)
	}
	if voidedSession.State != "VOIDED" {
		t.Fatalf("expected VOIDED, got %s", voidedSession.State)
	}
}

func TestManualFeeOverrideWhenNoRate(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	// No rate is seeded.
	seedShiftConfig(ctx, t, s, locationID)
	
	shift, err := s.GetOrCreateShift(ctx, locationID, 1, time.Now())
	if err != nil {
		t.Fatalf("get or create shift: %v", err)
	}

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
		Plate:       "B9999ZZZ",
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	manualFee := 15000.0
	checkOutAt := session.CheckInAt.Add(2 * time.Hour)
	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &manualFee,
	})
	if err != nil {
		t.Fatalf("checkout with manual fee: %v", err)
	}
	if session.FeeAmount == nil || *session.FeeAmount != manualFee {
		t.Fatalf("expected manual fee %v, got %v", manualFee, session.FeeAmount)
	}
}

func TestDuplicatePlateDetection(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	seedShiftConfig(ctx, t, s, locationID)
	
	shift, err := s.GetOrCreateShift(ctx, locationID, 1, time.Now())
	if err != nil {
		t.Fatalf("get or create shift: %v", err)
	}

	plate := "B7777XYZ"

	// First check-in.
	_, err = s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
		Plate:       plate,
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("first check-in: %v", err)
	}

	// Second check-in with same plate should succeed (duplicate detection is in handler).
	_, err = s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
		Plate:       plate,
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("second check-in should succeed: %v", err)
	}
}

func TestCheckInWithDifferentLocation(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	location1 := seedLocation(ctx, t, s)
	location2 := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, location1, operatorID)
	seedRate(ctx, t, s, location2, operatorID)
	seedShiftConfig(ctx, t, s, location1)
	seedShiftConfig(ctx, t, s, location2)
	
	shift1, _ := s.GetOrCreateShift(ctx, location1, 1, time.Now())
	shift2, _ := s.GetOrCreateShift(ctx, location2, 1, time.Now())

	plate := "B5555ABC"

	// Check in at location 1.
	_, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  location1,
		OperatorID:  operatorID,
		ShiftID:     shift1.ID,
		Plate:       plate,
		CityCode:    "B",
		VehicleType: "MOTO",
	})
	if err != nil {
		t.Fatalf("check-in at location 1: %v", err)
	}

	// Check in same plate at location 2 should succeed (different location).
	_, err = s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  location2,
		OperatorID:  operatorID,
		ShiftID:     shift2.ID,
		Plate:       plate,
		CityCode:    "B",
		VehicleType: "MOTO",
	})
	if err != nil {
		t.Fatalf("check-in at location 2 should succeed: %v", err)
	}
}

func generateReceiptNumber() string {
	return fmt.Sprintf("RCP-%d-%05d", time.Now().UnixMilli(), time.Now().UnixNano()%100000)
}

func seedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
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
	// Shift configs are auto-created with location
	return loc.ID
}

func seedOperator(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	roleID := ensureOperatorRole(ctx, t, s)
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         "Test Operator",
		Email:        randomEmail(),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func ensureOperatorRole(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	var id string
	err := s.Pool().QueryRow(ctx, `
		INSERT INTO roles (name, permissions)
		VALUES ('operator', $1)
		ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		RETURNING id
	`, []string{"sessions:*", "payments:*", "shifts:*"}).Scan(&id)
	if err != nil {
		t.Fatalf("ensure operator role: %v", err)
	}
	return id
}

func randomEmail() string {
	return fmt.Sprintf("operator-%d@test.local", time.Now().UnixNano())
}

func seedRate(ctx context.Context, t *testing.T, s *store.Store, locationID, createdBy string) {
	t.Helper()
	_, err := s.CreateRate(ctx, store.CreateRateInput{
		LocationID:           locationID,
		VehicleType:          "CAR",
		FirstHourRate:        5000,
		SubsequentHourlyRate: 3000,
		DailyFlatRate:        50000,
		EffectiveFrom:        time.Now().UTC().Add(-24 * time.Hour),
		CreatedBy:            createdBy,
	})
	if err != nil {
		t.Fatalf("create rate: %v", err)
	}
}

func seedShiftConfig(ctx context.Context, t *testing.T, s *store.Store, locationID string) {
	t.Helper()
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
}
