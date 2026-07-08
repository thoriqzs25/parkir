package reports_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/store"
	"github.com/thoriqzs/PARKIR/backend/internal/testutil"
)

func TestDailyRevenueReport(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s, "operator")
	shift := seedShift(ctx, t, s, operatorID, locationID)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 5000, "CAR", false)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 3000, "MOTO", false)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 2000, "MOTO", true)

	dr := store.DateRange{
		DateFrom: time.Now().AddDate(0, 0, -1),
		DateTo:   time.Now().Add(1 * time.Hour),
	}

	rows, err := s.ReportDailyRevenue(ctx, locationID, dr, false)
	if err != nil {
		t.Fatalf("daily revenue: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one row")
	}
	if rows[0].TotalRevenue != 8000 {
		t.Fatalf("expected revenue 8000 (excl voided), got %.0f", rows[0].TotalRevenue)
	}
	if rows[0].TransactionCount != 2 {
		t.Fatalf("expected 2 transactions, got %d", rows[0].TransactionCount)
	}

	rowsWithVoided, err := s.ReportDailyRevenue(ctx, locationID, dr, true)
	if err != nil {
		t.Fatalf("daily revenue with voided: %v", err)
	}
	if len(rowsWithVoided) == 0 {
		t.Fatal("expected at least one row")
	}
	if rowsWithVoided[0].VoidedCount != 1 {
		t.Fatalf("expected 1 voided, got %d", rowsWithVoided[0].VoidedCount)
	}
}

func TestOccupancyReport(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s, "operator")
	shift := seedShift(ctx, t, s, operatorID, locationID)

	for i := 0; i < 5; i++ {
		_, err := s.CreateSession(ctx, store.CreateSessionInput{
			LocationID:  locationID,
			OperatorID:  operatorID,
			ShiftID:     shift.ID,
			Plate:       fmt.Sprintf("B%dXYZ", 1000+i),
			CityCode:    "B",
			VehicleType: "CAR",
		})
		if err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	dr := store.DateRange{
		DateFrom: time.Now().AddDate(0, 0, -1),
		DateTo:   time.Now().Add(1 * time.Hour),
	}

	rows, err := s.ReportOccupancy(ctx, locationID, dr, "day")
	if err != nil {
		t.Fatalf("occupancy: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one row")
	}
	if rows[0].Count < 5 {
		t.Fatalf("expected at least 5 sessions, got %d", rows[0].Count)
	}
}

func TestVehicleBreakdownReport(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s, "operator")
	shift := seedShift(ctx, t, s, operatorID, locationID)

	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 5000, "CAR", false)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 3000, "MOTO", false)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 10000, "TRUCK", false)

	dr := store.DateRange{
		DateFrom: time.Now().AddDate(0, 0, -1),
		DateTo:   time.Now().Add(1 * time.Hour),
	}

	rows, err := s.ReportVehicleBreakdown(ctx, locationID, dr)
	if err != nil {
		t.Fatalf("vehicle breakdown: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 vehicle types, got %d", len(rows))
	}
}

func TestOperatorActivityReport(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s, "operator")
	shift := seedShift(ctx, t, s, operatorID, locationID)

	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 5000, "CAR", false)
	seedTransaction(ctx, t, s, locationID, operatorID, shift.ID, 3000, "MOTO", false)

	dr := store.DateRange{
		DateFrom: time.Now().AddDate(0, 0, -1),
		DateTo:   time.Now().Add(1 * time.Hour),
	}

	rows, err := s.ReportOperatorActivity(ctx, locationID, dr, "")
	if err != nil {
		t.Fatalf("operator activity: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one operator")
	}
	if rows[0].TotalRevenue != 8000 {
		t.Fatalf("expected revenue 8000, got %.0f", rows[0].TotalRevenue)
	}
}

func TestNinetyDayEnforcement(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := testutil.Ctx()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)

	dr := store.DateRange{
		DateFrom: time.Now().AddDate(0, 0, -100),
		DateTo:   time.Now(),
	}

	rows, err := s.ReportDailyRevenue(ctx, locationID, dr, false)
	if err != nil {
		t.Fatalf("daily revenue with 100-day range: %v", err)
	}
	// Should not error — enforcement is silent
	_ = rows
}

// Helpers

func seedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	loc, err := s.CreateLocation(ctx, store.CreateLocationInput{
		Name: fmt.Sprintf("Test Loc %d", time.Now().UnixNano()),
		Code: fmt.Sprintf("TL%d", time.Now().UnixNano()%10000),
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc.ID
}

func seedOperator(ctx context.Context, t *testing.T, s *store.Store, roleName string) string {
	t.Helper()
	var roleID string
	err := s.Pool().QueryRow(ctx, `
		INSERT INTO roles (name, permissions) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		RETURNING id
	`, roleName, []string{"*"}).Scan(&roleID)
	if err != nil {
		t.Fatalf("ensure role: %v", err)
	}
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         fmt.Sprintf("Op %d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("op-%d@test.local", time.Now().UnixNano()),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func seedShift(ctx context.Context, t *testing.T, s *store.Store, operatorID, locationID string) *store.Shift {
	t.Helper()
	shift, err := s.StartShift(ctx, store.StartShiftInput{
		OperatorID: operatorID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("start shift: %v", err)
	}
	return shift
}

func seedTransaction(ctx context.Context, t *testing.T, s *store.Store, locationID, operatorID, shiftID string, fee float64, vehicleType string, voided bool) {
	t.Helper()

	session, err := s.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  locationID,
		OperatorID:  operatorID,
		ShiftID:     shiftID,
		Plate:       fmt.Sprintf("B%dABC", time.Now().UnixNano()%100000),
		CityCode:    "B",
		VehicleType: vehicleType,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	session, err = s.UpdateSessionToPendingPayment(ctx, session.ID, store.CheckOutSessionInput{
		CheckOutAt: time.Now(),
		FeeAmount:  &fee,
	})
	if err != nil {
		t.Fatalf("checkout session: %v", err)
	}

	recNum := fmt.Sprintf("RPT-%d-%05d", time.Now().UnixMilli(), time.Now().UnixNano()%100000)
	tx, err := s.CreateTransaction(ctx, store.CreateTransactionInput{
		SessionID:            session.ID,
		LocationID:           locationID,
		ShiftID:              shiftID,
		OperatorID:           operatorID,
		VehicleType:          vehicleType,
		Plate:                session.Plate,
		CheckInAt:            session.CheckInAt,
		CheckOutAt:           time.Now(),
		DurationHours:        1,
		RateFirstHour:        fee,
		RateSubsequentHourly: fee,
		RateDaily:            fee,
		FeeAmount:            fee,
		PaymentMethod:        "CASH",
		AmountTendered:       &fee,
		ChangeAmount:         float64Ptr(0),
		ReceiptNumber:        recNum,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	if voided {
		_, err = s.VoidTransaction(ctx, tx.ID, operatorID, "test void")
		if err != nil {
			t.Fatalf("void transaction: %v", err)
		}
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}
