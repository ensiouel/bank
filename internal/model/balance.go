package model

import "github.com/google/uuid"

type Balance struct {
	UserID  uuid.UUID `db:"user_id"`
	Balance int64     `db:"balance"`
}
