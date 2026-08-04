package footprint

import "errors"

var (
	ErrJourneyNotFound = errors.New("footprint journey not found")
	ErrInvalidInput    = errors.New("invalid footprint input")
	ErrInvalidTrackURL = errors.New("invalid footprint track URL")
	ErrAlbumNotFound   = errors.New("footprint album not found")
)
