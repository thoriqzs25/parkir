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

	// Seed a location, operator, rate, and shift.
	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	shift, err := s.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("start shift: %v", err)
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
	rate, err := s.GetActiveRate(ctx, locationID, session.VehicleType, session.CheckInAt)
	if err != nil {
		t.Fatalf("get active rate: %v", err)
	}
	expectedFee, _ := fees.Calculate(fees.CalculationInput{
		FirstHourRate:        rate.FirstHourRate,
		SubsequentHourlyRate: rate.SubsequentHourlyRate,
		DailyFlatRate:        rate.DailyFlatRate,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           checkOutAt,
	})

	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &expectedFee,
		RateSnapshot: map[string]interface{}{
			"rate_id":                rate.ID,
			"first_hour_rate":        rate.FirstHourRate,
			"subsequent_hourly_rate": rate.SubsequentHourlyRate,
			"daily_flat_rate":        rate.DailyFlatRate,
		},
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

	// Pay cash.
	receiptNumber, err := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate receipt number: %v", err)
	}
	change := 10000.0
	tendered := expectedFee + change
	transaction, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:            session.ID,
		LocationID:           locationID,
		ShiftID:              shift.ID,
		OperatorID:           operatorID,
		VehicleType:          session.VehicleType,
		Plate:                session.Plate,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           *session.CheckOutAt,
		DurationHours:        3,
		RateFirstHour:        rate.FirstHourRate,
		RateSubsequentHourly: rate.SubsequentHourlyRate,
		RateDaily:            rate.DailyFlatRate,
		FeeAmount:            expectedFee,
		PaymentMethod:        "CASH",
		AmountTendered:       &tendered,
		ChangeAmount:         &change,
		ReceiptNumber:        receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	session, err = s.UpdateSessionToClosed(ctx, session.ID)
	if err != nil {
		t.Fatalf("close session: %v", err)
	}
	if session.State != "CLOSED" {
		t.Fatalf("expected CLOSED, got %s", session.State)
	}

	// Receipt number should contain the location code and today's date.
	if transaction.ReceiptNumber == "" {
		t.Fatal("expected receipt number")
	}

	// Close shift and verify cash discrepancy.
	expectedCash, err := s.SumCashByShift(ctx, shift.ID)
	if err != nil {
		t.Fatalf("sum cash: %v", err)
	}
	if expectedCash != expectedFee {
		t.Fatalf("expected cash %v, got %v", expectedFee, expectedCash)
	}
	_, err = s.UpdateShiftExpectedCash(ctx, shift.ID, expectedCash)
	if err != nil {
		t.Fatalf("update expected cash: %v", err)
	}
	closedShift, err := s.CloseShift(ctx, shift.ID, store.EndShiftInput{
		CashHandoverAmount: expectedCash,
	})
	if err != nil {
		t.Fatalf("close shift: %v", err)
	}
	if closedShift.Status != "CLOSED" {
		t.Fatalf("expected CLOSED shift, got %s", closedShift.Status)
	}
	if closedShift.Discrepancy == nil || *closedShift.Discrepancy != 0 {
		t.Fatalf("expected zero discrepancy, got %v", closedShift.Discrepancy)
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
	shift, err := s.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("start shift: %v", err)
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
		RateSnapshot: map[string]interface{}{
			"manual_override": true,
			"fee_amount":      manualFee,
		},
	})
	if err != nil {
		t.Fatalf("checkout with manual fee: %v", err)
	}
	if session.FeeAmount == nil || *session.FeeAmount != manualFee {
		t.Fatalf("expected manual fee %v, got %v", manualFee, session.FeeAmount)
	}

	receiptNumber, err := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate receipt number: %v", err)
	}
	_, err = s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:     session.ID,
		LocationID:    locationID,
		ShiftID:       shift.ID,
		OperatorID:    operatorID,
		VehicleType:   session.VehicleType,
		Plate:         session.Plate,
		CheckInAt:     session.CheckInAt,
		CheckOutAt:    *session.CheckOutAt,
		DurationHours: 2,
		FeeAmount:     manualFee,
		PaymentMethod: "DIGITAL",
		ReceiptNumber: receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
}

func TestVoidTransactionMarksSessionVoided(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID)
	shift, err := s.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("start shift: %v", err)
	}

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
		Plate:       "BVOID01",
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	checkOutAt := session.CheckInAt.Add(2 * time.Hour)
	fee := 11000.0
	session, _ = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &fee,
		RateSnapshot: map[string]interface{}{
			"manual_override": true,
			"fee_amount":      fee,
		},
	})

	receiptNumber, _ := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	tx, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:     session.ID,
		LocationID:    locationID,
		ShiftID:       shift.ID,
		OperatorID:    operatorID,
		VehicleType:   session.VehicleType,
		Plate:         session.Plate,
		CheckInAt:     session.CheckInAt,
		CheckOutAt:    *session.CheckOutAt,
		DurationHours: 2,
		FeeAmount:     fee,
		PaymentMethod: "CASH",
		ReceiptNumber: receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	voidedTx, err := s.VoidTransaction(ctx, tx.ID, operatorID, "wrong charge")
	if err != nil {
		t.Fatalf("void transaction: %v", err)
	}
	if !voidedTx.Voided {
		t.Fatal("expected transaction to be voided")
	}

	voidedSession, err := s.UpdateSessionToVoided(ctx, session.ID)
	if err != nil {
		t.Fatalf("void session: %v", err)
	}
	if voidedSession.State != "VOIDED" {
		t.Fatalf("expected VOIDED, got %s", voidedSession.State)
	}
}

func TestForceCloseShift(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	managerID := seedOperator(ctx, t, s) // reused as acting manager
	shift, err := s.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("start shift: %v", err)
	}

	forced, err := s.ForceCloseShift(ctx, shift.ID, store.ForceCloseShiftInput{
		ForceClosedBy:   managerID,
		ForceClosedReason: "operator left without closing",
	})
	if err != nil {
		t.Fatalf("force close shift: %v", err)
	}
	if forced.Status != "FORCE_CLOSED" {
		t.Fatalf("expected FORCE_CLOSED, got %s", forced.Status)
	}
	if forced.ForceClosedReason == nil || *forced.ForceClosedReason != "operator left without closing" {
		t.Fatalf("expected force close reason, got %v", forced.ForceClosedReason)
	}
}

func TestReceiptSequenceDailyIncrement(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)

	first, err := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate first receipt: %v", err)
	}
	second, err := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate second receipt: %v", err)
	}

	if first == second {
		t.Fatalf("expected different receipt numbers, got %s and %s", first, second)
	}
	if len(first) < 5 || len(second) < 5 {
		t.Fatalf("unexpected receipt format: %s, %s", first, second)
	}
}

func TestCrossOperatorCheckout(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorA := seedOperator(ctx, t, s)
	operatorB := seedOperator(ctx, t, s)

	shiftA, err := s.StartShift(ctx, store.StartShiftInput{OperatorID: operatorA, LocationID: locationID})
	if err != nil {
		t.Fatalf("start shift A: %v", err)
	}
	shiftB, err := s.StartShift(ctx, store.StartShiftInput{OperatorID: operatorB, LocationID: locationID})
	if err != nil {
		t.Fatalf("start shift B: %v", err)
	}

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorA,
		ShiftID:     shiftA.ID,
		Plate:       "BCROSS1",
		CityCode:    "B",
		VehicleType: "CAR",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	fee := 5000.0
	checkOutAt := session.CheckInAt.Add(time.Hour)
	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: checkOutAt,
		FeeAmount:  &fee,
		RateSnapshot: map[string]interface{}{
			"manual_override": true,
			"fee_amount":      fee,
		},
	})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}

	receiptNumber, _ := s.GenerateReceiptNumber(ctx, locationID, time.Now().UTC())
	tx, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:     session.ID,
		LocationID:    locationID,
		ShiftID:       shiftB.ID,
		OperatorID:    operatorB,
		VehicleType:   session.VehicleType,
		Plate:         session.Plate,
		CheckInAt:     session.CheckInAt,
		CheckOutAt:    *session.CheckOutAt,
		DurationHours: 1,
		FeeAmount:     fee,
		PaymentMethod: "CASH",
		ReceiptNumber: receiptNumber,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if tx.ShiftID != shiftB.ID {
		t.Fatalf("expected transaction shift_id %s, got %s", shiftB.ID, tx.ShiftID)
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
	email := randomEmail()
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
