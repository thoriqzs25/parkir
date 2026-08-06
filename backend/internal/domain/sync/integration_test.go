package sync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/store"
	"github.com/thoriqzs/PARKIR/backend/internal/testutil"
)

func TestOfflineSessionSync(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	seedShiftConfig(ctx, t, s, locationID)

	offlineSessionID := generateUUID()
	session, err := s.CreateOfflineSession(ctx, store.CreateOfflineSessionInput{
		ID:          offlineSessionID,
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftNumber: 1,
		Plate:       "BOFFLINE1",
		CityCode:    "B",
		VehicleType: "CAR",
		CheckInAt:   time.Now().UTC().Add(-2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create offline session: %v", err)
	}
	if session.ID != offlineSessionID {
		t.Fatalf("expected session id %s, got %s", offlineSessionID, session.ID)
	}
	if !session.OfflineSync {
		t.Fatal("expected offline_sync to be true")
	}
}

func TestOfflinePaymentSync(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	seedShiftConfig(ctx, t, s, locationID)

	sessionID := generateUUID()
	_, err := s.CreateOfflineSession(ctx, store.CreateOfflineSessionInput{
		ID:          sessionID,
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftNumber: 1,
		Plate:       "BOFFPAY1",
		CityCode:    "B",
		VehicleType: "CAR",
		CheckInAt:   time.Now().UTC().Add(-2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create offline session: %v", err)
	}

	fee := 8000.0
	checkOutAt := time.Now().UTC()
	_, err = s.UpdateSessionToPendingPayment(ctx, sessionID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &fee,
		RateSnapshot: map[string]interface{}{
			"manual_override": true,
			"fee_amount":      fee,
		},
	})
	if err != nil {
		t.Fatalf("checkout offline session: %v", err)
	}

	receiptNumber, err := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate receipt: %v", err)
	}

	amountTendered := 10000.0
	change := 2000.0
	tx, err := s.CreateOfflineTransaction(ctx, store.CreateOfflineTransactionInput{
		ID:             generateUUID(),
		SessionID:      sessionID,
		OperatorID:     operatorID,
		DurationHours:  2,
		FeeAmount:      fee,
		PaymentMethod:  "CASH",
		AmountTendered: &amountTendered,
		ChangeAmount:   &change,
		ReceiptNumber:  receiptNumber,
	})
	if err != nil {
		t.Fatalf("create offline transaction: %v", err)
	}
	if tx.ReceiptNumber == "" {
		t.Fatal("expected receipt number")
	}

	session, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if session.State != "CLOSED" {
		t.Fatalf("expected CLOSED, got %s", session.State)
	}
}

func seedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	loc, err := s.CreateLocation(ctx, store.CreateLocationInput{
		Name: "Test Location",
		Code: "TST",
		City: "Jakarta",
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc.ID
}

func seedOperator(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	roleID := ensureOperatorRole(ctx, t, s)
	email := fmt.Sprintf("operator-%d@test.local", time.Now().UnixNano())
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         "Test Operator",
		Email:        email,
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
	`, []string{"sessions:*", "payments:*", "shift-configs:*"}).Scan(&id)
	if err != nil {
		t.Fatalf("ensure operator role: %v", err)
	}
	return id
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
	_, _ = s.CreateLocationShiftConfig(ctx, store.CreateLocationShiftConfigInput{
		LocationID:  locationID,
		ShiftCode:   "08-16",
		ShiftNumber: 1,
		StartTime:   "08:00:00",
		EndTime:     "16:00:00",
	})
}

var testUUIDCounter int

func generateUUID() string {
	testUUIDCounter++
	return fmt.Sprintf("11111111-2222-3333-4444-%012d", testUUIDCounter)
}
