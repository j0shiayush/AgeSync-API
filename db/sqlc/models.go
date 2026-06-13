package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID   int32            `json:"id"`
	Name string           `json:"name"`
	Dob  pgtype.Date      `json:"dob"`
}