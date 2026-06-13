// Package middleware provides reusable Fiber middleware for the application.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-Id"

// RequestID injects a unique X-Request-Id header into every HTTP response.
// If the client already sent an X-Request-Id header its value is preserved;
// otherwise a new UUID v4 is generated.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(headerRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Make the ID available to handlers and downstream middleware via locals.
		c.Locals(headerRequestID, requestID)

		// Propagate the ID in the response header.
		c.Set(headerRequestID, requestID)

		return c.Next()
	}
}