package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
)

// VerifyHeaderValue verifies that a header is present and the value is an expected value
