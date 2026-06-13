package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteUser(ctx context.Context, id int32) error
	GetUserByID(ctx context.Context, id int32) (User, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)

type CreateUserParams struct {
	Name string      `json:"name"`
	Dob  pgtype.Date `json:"dob"`
}

type UpdateUserParams struct {
	ID   int32       `json:"id"`
	Name string      `json:"name"`
	Dob  pgtype.Date `json:"dob"`
}

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}