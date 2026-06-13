// Package handler implements the HTTP transport layer.
// Each method maps 1-to-1 to a route defined in internal/routes.
package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"AgeSync-API/internal/models"
	"AgeSync-API/internal/service"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests for the /users resource.
type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
	logger   *zap.Logger
}

// NewUserHandler constructs a UserHandler with its dependencies injected.
func NewUserHandler(svc service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
		logger:   logger,
	}
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: formatValidationError(err),
		})
	}

	resp, err := h.svc.CreateUser(c.Context(), req)
	if err != nil {
		h.logger.Error("handler.CreateUser", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetUser handles GET /users/:id.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "id must be a positive integer",
		})
	}

	resp, err := h.svc.GetUser(c.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: "user not found",
			})
		}
		h.logger.Error("handler.GetUser", zap.Error(err), zap.Int("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to fetch user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "id must be a positive integer",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: formatValidationError(err),
		})
	}

	resp, err := h.svc.UpdateUser(c.Context(), int32(id), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: "user not found",
			})
		}
		h.logger.Error("handler.UpdateUser", zap.Error(err), zap.Int("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to update user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// DeleteUser handles DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "id must be a positive integer",
		})
	}

	if err := h.svc.DeleteUser(c.Context(), int32(id)); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: "user not found",
			})
		}
		h.logger.Error("handler.DeleteUser", zap.Error(err), zap.Int("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to delete user",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users?page=1&limit=10.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	pq := models.PaginationQuery{Page: page, Limit: limit}
	if err := h.validate.Struct(pq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: formatValidationError(err),
		})
	}

	users, err := h.svc.ListUsers(c.Context(), page, limit)
	if err != nil {
		h.logger.Error("handler.ListUsers", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to list users",
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// ─── private helpers ──────────────────────────────────────────────────────────

// parseIDParam extracts the ":id" route parameter and validates it is a positive integer.
func parseIDParam(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, fiber.ErrBadRequest
	}
	return id, nil
}

// formatValidationError converts a validator.ValidationErrors into a single
// human-readable string returned to the client.
func formatValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) && len(ve) > 0 {
		return ve[0].Field() + ": " + ve[0].Tag() + " validation failed"
	}
	return "validation failed"
}