package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type User struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	RoleID       string     `json:"role_id"`
	RoleName     string     `json:"role_name,omitempty"`
	Status       string     `json:"status"`
	LocationIDs  []string   `json:"location_ids,omitempty"`
	PasswordHash string     `json:"-"`
	PINHash      *string    `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateUserInput struct {
	Name        string
	Email       string
	PasswordHash string
	RoleID      string
	LocationIDs []string
}

type UpdateUserInput struct {
	Name        *string
	Email       *string
	RoleID      *string
	LocationIDs []string
	Status      *string
}

func (s *Store) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, role_id, status, created_at, updated_at
	`, input.Name, input.Email, input.PasswordHash, input.RoleID).Scan(
		&user.ID, &user.Name, &user.Email, &user.RoleID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	if len(input.LocationIDs) > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO user_role_locations (user_id, location_id)
			SELECT $1, unnest($2::uuid[])
			ON CONFLICT DO NOTHING
		`, user.ID, input.LocationIDs)
		if err != nil {
			return nil, fmt.Errorf("insert user locations: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.password_hash, u.role_id, r.name as role_name, u.status
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.email = $1 AND u.status = 'ACTIVE'
	`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.RoleID, &user.RoleName, &user.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.password_hash, u.pin_hash, u.role_id, r.name as role_name, u.status, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.PINHash, &user.RoleID, &user.RoleName,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	locationIDs, err := s.getUserLocationIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	user.LocationIDs = locationIDs

	return &user, nil
}

func (s *Store) ListUsers(ctx context.Context, limit, offset int) ([]User, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.name, u.email, u.role_id, r.name as role_name, u.status, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.RoleID, &user.RoleName,
			&user.Status, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		locationIDs, err := s.getUserLocationIDs(ctx, user.ID)
		if err != nil {
			return nil, 0, err
		}
		user.LocationIDs = locationIDs
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user rows: %w", err)
	}

	var total int
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	return users, total, nil
}

func (s *Store) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if input.LocationIDs != nil {
		_, err = tx.Exec(ctx, `DELETE FROM user_role_locations WHERE user_id = $1`, id)
		if err != nil {
			return nil, fmt.Errorf("delete user locations: %w", err)
		}

		if len(input.LocationIDs) > 0 {
			_, err = tx.Exec(ctx, `
				INSERT INTO user_role_locations (user_id, location_id)
				SELECT $1, unnest($2::uuid[])
				ON CONFLICT DO NOTHING
			`, id, input.LocationIDs)
			if err != nil {
				return nil, fmt.Errorf("insert user locations: %w", err)
			}
		}
	}

	var user User
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET
			name = COALESCE($2, name),
			email = COALESCE($3, email),
			role_id = COALESCE($4, role_id),
			status = COALESCE($5, status),
			updated_at = now()
		WHERE id = $1
		RETURNING id, name, email, role_id, status, created_at, updated_at
	`, id, input.Name, input.Email, input.RoleID, input.Status).Scan(
		&user.ID, &user.Name, &user.Email, &user.RoleID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	locationIDs, err := s.getUserLocationIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	user.LocationIDs = locationIDs

	return &user, nil
}

func (s *Store) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1
	`, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (s *Store) UpdatePIN(ctx context.Context, id string, pinHash string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE users SET pin_hash = $2, updated_at = now() WHERE id = $1
	`, id, pinHash)
	if err != nil {
		return fmt.Errorf("update pin: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (s *Store) DeactivateUser(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE users SET status = 'DEACTIVATED', updated_at = now() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (s *Store) getUserLocationIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT location_id FROM user_role_locations WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user locations: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan location id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
