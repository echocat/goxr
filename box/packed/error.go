package packed

import (
	"errors"
)

var (
	ErrInvalidHeaderVersion = errors.New("invalid header version")
	ErrActiveEntryWriter    = errors.New("there is another entry writer active and not closed")
)
