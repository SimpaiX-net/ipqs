package ipqs

import "errors"

var (
	ErrInvalidProtocol = errors.New("given protocol in rotating_proxy is not supported")
	ErrBadIPRep        = errors.New("bad ip reputation")
	ErrUnknown         = errors.New("unknown ip reputation")
)
