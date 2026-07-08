package store

import (
	"context"
	"fmt"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/timeutil"
)

// GenerateReceiptNumber atomically creates the next receipt number for the location on the given date.
// Format: [LOCATION_CODE]-[YYYYMMDD]-[SEQUENCE]
// The sequence is strictly sequential per location per day.
func (s *Store) GenerateReceiptNumber(ctx context.Context, locationID string, date time.Time) (string, error) {
	seqDate := timeutil.DateJakarta(date)
	dateStr := seqDate.Format("20060102")

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin receipt sequence tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var lastNumber int
	err = tx.QueryRow(ctx, `
		INSERT INTO receipt_sequences (location_id, sequence_date, last_number)
		VALUES ($1, $2, 1)
		ON CONFLICT (location_id, sequence_date)
		DO UPDATE SET last_number = receipt_sequences.last_number + 1
		RETURNING last_number
	`, locationID, seqDate).Scan(&lastNumber)
	if err != nil {
		return "", fmt.Errorf("next receipt sequence: %w", err)
	}

	var code string
	err = tx.QueryRow(ctx, `SELECT code FROM locations WHERE id = $1`, locationID).Scan(&code)
	if err != nil {
		return "", fmt.Errorf("get location code: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit receipt sequence tx: %w", err)
	}

	return fmt.Sprintf("%s-%s-%05d", code, dateStr, lastNumber), nil
}
