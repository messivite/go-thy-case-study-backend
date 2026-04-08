package chat

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrInvalidRole         = errors.New("invalid message role")
	ErrMissingContent      = errors.New("missing message content")
	ErrUnauthorized        = errors.New("not authorized")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidSessionID    = errors.New("invalid session id")
)
