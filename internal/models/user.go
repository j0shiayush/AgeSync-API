// Package models defines the domain models and HTTP transport DTOs used throughout the application.
package models

// CreateUserRequest is the JSON body accepted by POST /users.
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,datetime=2006-01-02"`
}

// UpdateUserRequest is the JSON body accepted by PUT /users/:id.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,datetime=2006-01-02"`
}

// UserResponse is returned by POST /users and PUT /users/:id (no age field).
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
}

// UserDetailResponse is returned by GET /users/:id and GET /users (includes age).
type UserDetailResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
	Age  int    `json:"age"`
}

// PaginationQuery holds validated pagination parameters parsed from the query string.
type PaginationQuery struct {
	Page  int `query:"page"  validate:"min=1"`
	Limit int `query:"limit" validate:"min=1,max=100"`
}

// ErrorResponse is the standard error envelope returned on failures.
type ErrorResponse struct {
	Error string `json:"error"`
}