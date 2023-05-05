package model

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID        uuid.UUID  `db:"id"`
	PayeeID   uuid.UUID  `db:"payee_id"`
	PayerID   *uuid.UUID `db:"payer_id"`
	Type      string     `db:"type"`
	Amount    int64      `db:"amount"`
	Comment   string     `db:"comment"`
	CreatedAt time.Time  `db:"created_at"`
}
