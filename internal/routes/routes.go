// Package routes wires Fiber routes to their handlers and registers global middleware.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"AgeSync-API/internal/handler"
	"AgeSync-API/internal/middleware"
	"go.uber.org/zap"
)

// Register attaches all application routes and middleware to the provided Fiber app.
func Register(app *fiber.App, userHandler *handler.UserHandler, logger *zap.Logger) {
	// ── Global middleware ─────────────────────────────────────────────────────
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(logger))

	// ── Health check ──────────────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	// ── User resource ─────────────────────────────────────────────────────────
	users := app.Group("/users")
	{
		users.Post("/", userHandler.CreateUser)
		users.Get("/", userHandler.ListUsers)
		users.Get("/:id", userHandler.GetUser)
		users.Put("/:id", userHandler.UpdateUser)
		users.Delete("/:id", userHandler.DeleteUser)
	}
}