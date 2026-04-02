package com

import "errors"

var (
	ErrFailedReadingHeader        = errors.New("failed to read MIME header")
	ErrInvalidContentLengthHeader = errors.New("invalid Content-Length header")
	ErrFailedDecode               = errors.New("failed to decode incoming request")
	ErrServerNotInitialized       = errors.New("client did not initialize server")
	ErrMethodNotFound             = errors.New("no handler for method found")
)
