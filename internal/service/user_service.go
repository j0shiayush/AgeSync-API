// Package service implements the business logic for the user domain.
// Age calculation lives here — it is a pure function of the current time and dob,
// which makes it trivially testable without any database involvement.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	db "AgeSync-API/db/sqlc"
	"AgeSync-API/internal/models"
	"AgeSync-API/internal/repository"
	"go.uber.org/zap"
)

const dateLayout = "2006-01-02"

// ErrUserNotFound is returned when a requested user does not exist in the database.
var ErrUserNotFound = errors.New("user not found")

// UserService defines the business-logic contract for the user domain.
type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetUser(ctx context.Context, id int32) (models.UserDetailResponse, error)
	UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error)
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, page, limit int) ([]models.UserDetailResponse, error)
}

// userService is the concrete implementation of UserService.
type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

// NewUserService constructs a UserService with its dependencies injected.
func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser validates, persists, and returns a new user.
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	s.logger.Info("service.CreateUser: creating user", zap.String("name", req.Name))

	resp, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.Error("service.CreateUser: repo error", zap.Error(err))
		return models.UserResponse{}, fmt.Errorf("service.CreateUser: %w", err)
	}

	s.logger.Info("service.CreateUser: user created", zap.Int32("id", resp.ID))
	return resp, nil
}

// GetUser fetches a user by ID and enriches the response with the calculated age.
func (s *userService) GetUser(ctx context.Context, id int32) (models.UserDetailResponse, error) {
	s.logger.Info("service.GetUser: fetching user", zap.Int32("id", id))

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return models.UserDetailResponse{}, ErrUserNotFound
		}
		s.logger.Error("service.GetUser: repo error", zap.Error(err), zap.Int32("id", id))
		return models.UserDetailResponse{}, fmt.Errorf("service.GetUser: %w", err)
	}

	return s.toDetailResponse(user), nil
}

// UpdateUser modifies an existing user and returns the updated resource.
func (s *userService) UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	s.logger.Info("service.UpdateUser: updating user", zap.Int32("id", id))

	// Verify the user exists before attempting an update.
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if isNotFound(err) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, fmt.Errorf("service.UpdateUser: existence check: %w", err)
	}

	resp, err := s.repo.Update(ctx, id, req)
	if err != nil {
		s.logger.Error("service.UpdateUser: repo error", zap.Error(err), zap.Int32("id", id))
		return models.UserResponse{}, fmt.Errorf("service.UpdateUser: %w", err)
	}

	s.logger.Info("service.UpdateUser: user updated", zap.Int32("id", id))
	return resp, nil
}

// DeleteUser removes a user by ID.
func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	s.logger.Info("service.DeleteUser: deleting user", zap.Int32("id", id))

	// Verify the user exists before deletion.
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if isNotFound(err) {
			return ErrUserNotFound
		}
		return fmt.Errorf("service.DeleteUser: existence check: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("service.DeleteUser: repo error", zap.Error(err), zap.Int32("id", id))
		return fmt.Errorf("service.DeleteUser: %w", err)
	}

	s.logger.Info("service.DeleteUser: user deleted", zap.Int32("id", id))
	return nil
}

// ListUsers returns a page of users, each enriched with a dynamically calculated age.
func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]models.UserDetailResponse, error) {
	s.logger.Info("service.ListUsers", zap.Int("page", page), zap.Int("limit", limit))

	users, err := s.repo.List(ctx, page, limit)
	if err != nil {
		s.logger.Error("service.ListUsers: repo error", zap.Error(err))
		return nil, fmt.Errorf("service.ListUsers: %w", err)
	}

	result := make([]models.UserDetailResponse, 0, len(users))
	for _, u := range users {
		result = append(result, s.toDetailResponse(u))
	}
	return result, nil
}

// ─── age calculation ──────────────────────────────────────────────────────────

// CalculateAge returns the integer age of a person born on dob relative to now.
//
// The calculation uses calendar arithmetic rather than simple year subtraction,
// so it correctly handles leap-year birthdays and pre-birthday months:
//   - If the person's birthday has already occurred this calendar year → age = now.Year - dob.Year
//   - If not yet occurred this year → age = now.Year - dob.Year - 1
//
// This function is exported so it can be unit-tested independently of any I/O.
func CalculateAge(dob, now time.Time) int {
	years := now.Year() - dob.Year()

	// Has the birthday already passed this year?
	dobThisYear := time.Date(now.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(dobThisYear) {
		years--
	}

	return years
}

// ─── private helpers ──────────────────────────────────────────────────────────

// toDetailResponse maps a db.User to a UserDetailResponse, calculating age in-process.
func (s *userService) toDetailResponse(u db.User) models.UserDetailResponse {
	dobStr := ""
	age := 0

	if u.Dob.Valid {
		dobStr = u.Dob.Time.UTC().Format(dateLayout)
		age = CalculateAge(u.Dob.Time.UTC(), time.Now().UTC())
	}

	return models.UserDetailResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  dobStr,
		Age:  age,
	}
}

// isNotFound returns true when the error originates from a pgx "no rows" result.
func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}