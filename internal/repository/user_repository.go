// Package repository provides database access for the user domain.
// It wraps the SQLC-generated Queries, exposing a clean interface to the service layer.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "AgeSync-API/db/sqlc"
	"AgeSync-API/internal/models"
)

const dateLayout = "2006-01-02"

// UserRepository defines the persistence contract for the user domain.
type UserRepository interface {
	Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetByID(ctx context.Context, id int32) (db.User, error)
	Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, page, limit int) ([]db.User, error)
}

// userRepository is the PostgreSQL-backed implementation of UserRepository.
type userRepository struct {
	q db.Querier
}

// NewUserRepository returns a UserRepository backed by the provided SQLC Querier.
func NewUserRepository(q db.Querier) UserRepository {
	return &userRepository{q: q}
}

// Create inserts a new user and returns the created resource representation.
func (r *userRepository) Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	dob, err := parseDateToPgtype(req.DOB)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("repository.Create: parse dob: %w", err)
	}

	user, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("repository.Create: db: %w", err)
	}

	return toUserResponse(user), nil
}

// GetByID fetches a single user row by primary key.
func (r *userRepository) GetByID(ctx context.Context, id int32) (db.User, error) {
	user, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return db.User{}, fmt.Errorf("repository.GetByID: db: %w", err)
	}
	return user, nil
}

// Update modifies an existing user row and returns the updated resource.
func (r *userRepository) Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	dob, err := parseDateToPgtype(req.DOB)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("repository.Update: parse dob: %w", err)
	}

	user, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:   id,
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("repository.Update: db: %w", err)
	}

	return toUserResponse(user), nil
}

// Delete removes a user row; the caller should check existence first.
func (r *userRepository) Delete(ctx context.Context, id int32) error {
	if err := r.q.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("repository.Delete: db: %w", err)
	}
	return nil
}

// List returns a paginated slice of user rows.
func (r *userRepository) List(ctx context.Context, page, limit int) ([]db.User, error) {
	offset := (page - 1) * limit
	users, err := r.q.ListUsers(ctx, db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("repository.List: db: %w", err)
	}
	return users, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

// parseDateToPgtype converts a "YYYY-MM-DD" string into a pgtype.Date.
func parseDateToPgtype(s string) (pgtype.Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("parseDateToPgtype: %w", err)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// toUserResponse maps a db.User row to a UserResponse DTO.
func toUserResponse(u db.User) models.UserResponse {
	return models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  formatDate(u.Dob),
	}
}

// formatDate converts a pgtype.Date to a "YYYY-MM-DD" string.
func formatDate(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.UTC().Format(dateLayout)
}