package domain

import (
	"bank/internal/model"
	"github.com/google/uuid"
	"time"
)

const (
	Transfer string = "transfer"
	Debet    string = "debet"
	Credit   string = "credit"
)

type Transaction struct {
	ID        uuid.UUID  `json:"id"`
	PayeeID   uuid.UUID  `json:"payee_id,omitempty"`
	PayerID   *uuid.UUID `json:"payer_id,omitempty"`
	Type      string     `json:"type"`
	Amount    float64    `json:"amount"`
	Comment   string     `json:"comment,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func TransactionFromModel(transaction model.Transaction) Transaction {
	return Transaction{
		ID:        transaction.ID,
		PayeeID:   transaction.PayeeID,
		PayerID:   transaction.PayerID,
		Type:      transaction.Type,
		Amount:    float64(transaction.Amount) / 100.0,
		Comment:   transaction.Comment,
		CreatedAt: transaction.CreatedAt,
	}
}

func TransactionsFromModels(transactions []model.Transaction) []Transaction {
	res := make([]Transaction, len(transactions))

	for i, transaction := range transactions {
		res[i] = TransactionFromModel(transaction)
	}

	return res
}
