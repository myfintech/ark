package mhttp

import (
	"net/http"
)

// RawMiddleware a raw middleware interface that honors the stdlib
type RawMiddleware func(handler http.Handler) http.Handler
