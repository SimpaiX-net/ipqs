package ipqs

import "errors"

var (
	ErrBadIPRep        = errors.New("bad ip reputation")
	ErrUnknown         = errors.New("unknown ip reputation")
)
