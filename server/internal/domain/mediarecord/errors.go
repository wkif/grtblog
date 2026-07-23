package mediarecord

import "errors"

var (
	ErrNotFound      = errors.New("media record not found")
	ErrInvalidStatus = errors.New("invalid media record status")
	ErrInvalidType   = errors.New("invalid media record type")
)
