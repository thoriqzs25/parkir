package gate_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/store"
	"github.com/thoriqzs/PARKIR/backend/internal/testutil"
)

func TestGetGateInfo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)
	operatorID := seedOperator(ctx, t, s)
	seedRate(ctx, t, s, locationID, operatorID, "CAR", -24*time.Hour, nil)
	seedRate(ctx, t, s, locationID, operatorID, "MOTO", 7*24*time.Hour, nil)

	info, err := s.GetGateInfo(ctx, locationID)
	if err != nil {
		t.Fatalf("GetGateInfo: %v", err)
	}
	if info.Location.Name == "" {
		t.Fatal("expected location name to be non-empty")
	}
	if len(info.Rates) != 1 {
		t.Fatalf("expected 1 active rate (future rate excluded), got %d", len(info.Rates))
	}
	if info.Rates[0].VehicleType != "CAR" {
		t.Fatalf("expected vehicle type CAR, got %s", info.Rates[0].VehicleType)
	}
}

func TestGetGateInfo_NoRates(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)

	info, err := s.GetGateInfo(ctx, locationID)
	if err != nil {
		t.Fatalf("GetGateInfo: %v", err)
	}
	if info.Location.Name == "" {
		t.Fatal("expected location name to be non-empty")
	}
	if info.Rates == nil {
		t.Fatal("expected rates to be non-nil (empty slice)")
	}
	if len(info.Rates) != 0 {
		t.Fatalf("expected 0 rates, got %d", len(info.Rates))
	}
}

func TestGetGateInfo_LocationNotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	_, err := s.GetGateInfo(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent location")
	}
}

func TestRegisterGate(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	locationID := seedLocation(ctx, t, s)

	gate, err := s.RegisterGate(ctx, store.RegisterGateInput{
		DeviceID:   "gate-device-001",
		Name:       "Gate A",
		LocationID: &locationID,
		IPAddress:  "192.168.1.100",
	})
	if err != nil {
		t.Fatalf("RegisterGate: %v", err)
	}
	if gate.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if gate.DeviceID != "gate-device-001" {
		t.Fatalf("expected device_id gate-device-001, got %s", gate.DeviceID)
	}
	if gate.LocationID == nil || *gate.LocationID != locationID {
		t.Fatalf("expected location_id %s, got %v", locationID, gate.LocationID)
	}
}

func TestRegisterGate_DuplicateDeviceID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	_, err := s.RegisterGate(ctx, store.RegisterGateInput{
		DeviceID: "dup-device",
		Name:     "First",
	})
	if err != nil {
		t.Fatalf("first register: %v", err)
	}

	_, err = s.RegisterGate(ctx, store.RegisterGateInput{
		DeviceID: "dup-device",
		Name:     "Second",
	})
	if err == nil {
		t.Fatal("expected error for duplicate device_id")
	}
}

func TestListGates(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	loc1 := seedLocation(ctx, t, s)
	loc2 := seedLocation(ctx, t, s)

	s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "g1", Name: "Gate 1", LocationID: &loc1})
	s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "g2", Name: "Gate 2", LocationID: &loc1})
	s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "g3", Name: "Gate 3", LocationID: &loc2})

	gates, err := s.ListGates(ctx, loc1)
	if err != nil {
		t.Fatalf("ListGates: %v", err)
	}
	if len(gates) != 2 {
		t.Fatalf("expected 2 gates in loc1, got %d", len(gates))
	}

	allGates, err := s.ListGates(ctx, "")
	if err != nil {
		t.Fatalf("ListGates (all): %v", err)
	}
	if len(allGates) != 3 {
		t.Fatalf("expected 3 total gates, got %d", len(allGates))
	}
}

func TestUpdateGate(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	gate, err := s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "update-test", Name: "Old Name"})
	if err != nil {
		t.Fatalf("RegisterGate: %v", err)
	}

	newName := "New Name"
	updated, err := s.UpdateGate(ctx, gate.ID, store.UpdateGateInput{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateGate: %v", err)
	}
	if updated.Name != "New Name" {
		t.Fatalf("expected name 'New Name', got %s", updated.Name)
	}

	fetched, err := s.GetGateByID(ctx, gate.ID)
	if err != nil {
		t.Fatalf("GetGateByID: %v", err)
	}
	if fetched.Name != "New Name" {
		t.Fatalf("expected fetched name 'New Name', got %s", fetched.Name)
	}
}

func TestDeleteGate(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	gate, err := s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "delete-test", Name: "To Delete"})
	if err != nil {
		t.Fatalf("RegisterGate: %v", err)
	}

	err = s.DeleteGate(ctx, gate.ID)
	if err != nil {
		t.Fatalf("DeleteGate: %v", err)
	}

	_, err = s.GetGateByID(ctx, gate.ID)
	if err == nil {
		t.Fatal("expected ErrNotFound after delete")
	}
}

func TestGetGateByDeviceID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup()

	ctx := context.Background()
	s := tdb.Store

	gate, err := s.RegisterGate(ctx, store.RegisterGateInput{DeviceID: "find-by-did", Name: "Device Lookup"})
	if err != nil {
		t.Fatalf("RegisterGate: %v", err)
	}

	found, err := s.GetGateByDeviceID(ctx, "find-by-did")
	if err != nil {
		t.Fatalf("GetGateByDeviceID: %v", err)
	}
	if found.ID != gate.ID {
		t.Fatalf("expected ID %s, got %s", gate.ID, found.ID)
	}

	_, err = s.GetGateByDeviceID(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected ErrNotFound for nonexistent device_id")
	}
}

// Helpers

func seedLocation(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	loc, err := s.CreateLocation(ctx, store.CreateLocationInput{
		Name:    fmt.Sprintf("Gate Test Loc %d", time.Now().UnixNano()),
		Code:    fmt.Sprintf("GT-%d", time.Now().UnixNano()%100000),
		Address: "Test Address",
		City:    "Test City",
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc.ID
}

func seedOperator(ctx context.Context, t *testing.T, s *store.Store) string {
	t.Helper()
	var roleID string
	err := s.Pool().QueryRow(ctx, `
		INSERT INTO roles (name, permissions)
		VALUES ('gate-test-op', $1)
		ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		RETURNING id
	`, []string{"*"}).Scan(&roleID)
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	user, err := s.CreateUser(ctx, store.CreateUserInput{
		Name:         "Gate Test Operator",
		Email:        fmt.Sprintf("gate-op-%d@test.local", time.Now().UnixNano()),
		PasswordHash: "hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func seedRate(ctx context.Context, t *testing.T, s *store.Store, locationID, createdBy, vehicleType string, effectiveFromOffset time.Duration, effectiveUntil *time.Time) {
	t.Helper()
	_, err := s.CreateRate(ctx, store.CreateRateInput{
		LocationID:           locationID,
		VehicleType:          vehicleType,
		FirstHourRate:        5000,
		SubsequentHourlyRate: 3000,
		DailyFlatRate:        50000,
		EffectiveFrom:        time.Now().UTC().Add(effectiveFromOffset),
		EffectiveUntil:       effectiveUntil,
		CreatedBy:            createdBy,
	})
	if err != nil {
		t.Fatalf("create rate: %v", err)
	}
}
