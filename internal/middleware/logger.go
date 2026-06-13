package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Logger returns a Fiber middleware that logs the method, path, status code,
// latency, and request ID for every incoming HTTP request using Uber Zap.
func Logger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue to the next handler.
		err := c.Next()

		latency := time.Since(start)
		requestID, _ := c.Locals(headerRequestID).(string)

		// Choose log level based on status code.
		status := c.Response().StatusCode()
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Get(fiber.HeaderUserAgent)),
		}

		switch {
		case status >= 500:
			log.Error("request completed", fields...)
		case status >= 400:
			log.Warn("request completed", fields...)
		default:
			log.Info("request completed", fields...)
		}

		return err
	}
}