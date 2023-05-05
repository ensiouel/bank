package domain

import (
	"bank/internal/model"
	"github.com/google/uuid"
)

type Balance struct {
	UserID  uuid.UUID `json:"user_id"`
	Balance float64   `json:"balance"`
}

func BalanceFromModel(balance model.Balance) Balance {
	return Balance{
		UserID:  balance.UserID,
		Balance: float64(balance.Balance) / 100.0,
	}
}
