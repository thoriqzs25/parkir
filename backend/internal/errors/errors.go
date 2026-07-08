package errors

import "errors"

var (
	ErrNotFound              = errors.New("resource not found")
	ErrConflict              = errors.New("resource conflict")
	ErrInvalidInput          = errors.New("invalid input")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrRateOverlap           = errors.New("rate overlap detected")
	ErrInvalidPermission     = errors.New("invalid permission")
	ErrFinancePermission     = errors.New("only owners can grant finance permissions")
	ErrInvalidState          = errors.New("invalid state")
	ErrShiftLocationMismatch = errors.New("shift location mismatch")
	ErrInsufficientPayment   = errors.New("insufficient payment")
)
