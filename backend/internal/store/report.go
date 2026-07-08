package store

import (
	"context"
	"fmt"
	"time"
)

type DailyRevenueRow struct {
	Date             time.Time `json:"date"`
	TotalRevenue     float64   `json:"total_revenue"`
	TransactionCount int       `json:"transaction_count"`
	AverageFee       float64   `json:"average_fee"`
	VoidedCount      int       `json:"voided_count"`
	VoidedAmount     float64   `json:"voided_amount"`
}

type OccupancyRow struct {
	Bucket time.Time `json:"bucket"`
	Count  int       `json:"count"`
}

type VehicleBreakdownRow struct {
	VehicleType string  `json:"vehicle_type"`
	Count       int     `json:"count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type OperatorActivityRow struct {
	OperatorID   string  `json:"operator_id"`
	OperatorName string  `json:"operator_name"`
	SessionCount int     `json:"session_count"`
	TotalRevenue float64 `json:"total_revenue"`
	ShiftHours   float64 `json:"shift_hours"`
}

type DateRange struct {
	DateFrom time.Time
	DateTo   time.Time
}

func (s *Store) ReportDailyRevenue(ctx context.Context, locationID string, dr DateRange, includeVoided bool) ([]DailyRevenueRow, error) {
	if dr.DateTo.Sub(dr.DateFrom) > 90*24*time.Hour {
		dr.DateTo = dr.DateFrom.Add(90 * 24 * time.Hour)
	}

	query := `
		SELECT
			t.created_at::date AS date,
			COALESCE(SUM(t.fee_amount) FILTER (WHERE t.voided = false), 0) AS total_revenue,
			COUNT(*) FILTER (WHERE t.voided = false) AS transaction_count,
			COALESCE(AVG(t.fee_amount) FILTER (WHERE t.voided = false), 0) AS average_fee,
			COUNT(*) FILTER (WHERE t.voided = true) AS voided_count,
			COALESCE(SUM(t.fee_amount) FILTER (WHERE t.voided = true), 0) AS voided_amount
		FROM transactions t
		WHERE t.location_id = $1
		  AND t.created_at::date >= $2::date
		  AND t.created_at::date <= $3::date
		GROUP BY t.created_at::date
		ORDER BY t.created_at::date ASC
	`

	rows, err := s.pool.Query(ctx, query, locationID, dr.DateFrom, dr.DateTo)
	if err != nil {
		return nil, fmt.Errorf("report daily revenue: %w", err)
	}
	defer rows.Close()

	var results []DailyRevenueRow
	for rows.Next() {
		var r DailyRevenueRow
		if err := rows.Scan(&r.Date, &r.TotalRevenue, &r.TransactionCount, &r.AverageFee, &r.VoidedCount, &r.VoidedAmount); err != nil {
			return nil, fmt.Errorf("scan daily revenue: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) ReportOccupancy(ctx context.Context, locationID string, dr DateRange, granularity string) ([]OccupancyRow, error) {
	if dr.DateTo.Sub(dr.DateFrom) > 90*24*time.Hour {
		dr.DateTo = dr.DateFrom.Add(90 * 24 * time.Hour)
	}

	var bucketExpr string
	if granularity == "hour" {
		bucketExpr = "date_trunc('hour', s.check_in_at)"
	} else {
		bucketExpr = "date_trunc('day', s.check_in_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s AS bucket, COUNT(*)::int AS count
		FROM sessions s
		WHERE s.location_id = $1
		  AND s.check_in_at >= $2
		  AND s.check_in_at <= $3
		GROUP BY bucket
		ORDER BY bucket ASC
	`, bucketExpr)

	rows, err := s.pool.Query(ctx, query, locationID, dr.DateFrom, dr.DateTo)
	if err != nil {
		return nil, fmt.Errorf("report occupancy: %w", err)
	}
	defer rows.Close()

	var results []OccupancyRow
	for rows.Next() {
		var r OccupancyRow
		if err := rows.Scan(&r.Bucket, &r.Count); err != nil {
			return nil, fmt.Errorf("scan occupancy: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) ReportVehicleBreakdown(ctx context.Context, locationID string, dr DateRange) ([]VehicleBreakdownRow, error) {
	if dr.DateTo.Sub(dr.DateFrom) > 90*24*time.Hour {
		dr.DateTo = dr.DateFrom.Add(90 * 24 * time.Hour)
	}

	query := `
		SELECT
			t.vehicle_type,
			COUNT(*)::int AS count,
			COALESCE(SUM(t.fee_amount), 0) AS total_revenue
		FROM transactions t
		WHERE t.location_id = $1
		  AND t.created_at::date >= $2::date
		  AND t.created_at::date <= $3::date
		  AND t.voided = false
		GROUP BY t.vehicle_type
		ORDER BY t.vehicle_type ASC
	`

	rows, err := s.pool.Query(ctx, query, locationID, dr.DateFrom, dr.DateTo)
	if err != nil {
		return nil, fmt.Errorf("report vehicle breakdown: %w", err)
	}
	defer rows.Close()

	var results []VehicleBreakdownRow
	for rows.Next() {
		var r VehicleBreakdownRow
		if err := rows.Scan(&r.VehicleType, &r.Count, &r.TotalRevenue); err != nil {
			return nil, fmt.Errorf("scan vehicle breakdown: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) ReportOperatorActivity(ctx context.Context, locationID string, dr DateRange, operatorID string) ([]OperatorActivityRow, error) {
	if dr.DateTo.Sub(dr.DateFrom) > 90*24*time.Hour {
		dr.DateTo = dr.DateFrom.Add(90 * 24 * time.Hour)
	}

	whereExtra := ""
	args := []interface{}{locationID, dr.DateFrom, dr.DateTo}
	argIdx := 3

	if operatorID != "" {
		argIdx++
		whereExtra = fmt.Sprintf(" AND t.operator_id = $%d", argIdx)
		args = append(args, operatorID)
	}

	query := fmt.Sprintf(`
		SELECT
			t.operator_id,
			COALESCE(u.name, 'Unknown') AS operator_name,
			COUNT(DISTINCT t.session_id)::int AS session_count,
			COALESCE(SUM(t.fee_amount), 0) AS total_revenue,
			COALESCE(EXTRACT(EPOCH FROM (sh.ended_at - sh.started_at)) / 3600, 0) AS shift_hours
		FROM transactions t
		JOIN users u ON u.id = t.operator_id
		LEFT JOIN shifts sh ON sh.id = t.shift_id AND sh.operator_id = t.operator_id
		WHERE t.location_id = $1
		  AND t.created_at::date >= $2::date
		  AND t.created_at::date <= $3::date
		  AND t.voided = false
		  %s
		GROUP BY t.operator_id, u.name, sh.ended_at, sh.started_at
		ORDER BY total_revenue DESC
	`, whereExtra)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("report operator activity: %w", err)
	}
	defer rows.Close()

	var results []OperatorActivityRow
	for rows.Next() {
		var r OperatorActivityRow
		if err := rows.Scan(&r.OperatorID, &r.OperatorName, &r.SessionCount, &r.TotalRevenue, &r.ShiftHours); err != nil {
			return nil, fmt.Errorf("scan operator activity: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}