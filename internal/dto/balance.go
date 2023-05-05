package dto

import (
	"bank/pkg/sort"
	"github.com/google/uuid"
)

type SelectTransactionFilter struct {
	Sort  string `query:"sort"`
	Order string `query:"order"`
	sort.Pagination
}

func (filter SelectTransactionFilter) GetSort() string {
	switch filter.Sort {
	case "amount", "created_at":
		return filter.Sort
	}

	return "created_at"
}

func (filter SelectTransactionFilter) GetOrder() string {
	switch filter.Order {
	case "asc", "desc":
		return filter.Order
	}

	return "asc"
}

type GetBalance struct {
	UserID   uuid.UUID `query:"user_id"`
	Currency string    `query:"currency"`
}

type SelectTransaction struct {
	UserID uuid.UUID `query:"user_id"`
	SelectTransactionFilter
}

type Transfer struct {
	PayeeID uuid.UUID `json:"payee_id" example:"a39c71f8-6d1b-466c-8367-ebd86764268b"`
	PayerID uuid.UUID `json:"payer_id" example:"bcf4b5f7-8f73-4205-82e6-cf20e898a98a"`
	Amount  float64   `json:"amount" validate:"gt=0" example:"100"`
	Comment string    `json:"comment" example:"paid the debt"`
}

type Debet struct {
	UserID  uuid.UUID `json:"user_id" example:"bcf4b5f7-8f73-4205-82e6-cf20e898a98a"`
	Amount  float64   `json:"amount" validate:"gt=0" example:"100"`
	Comment string    `json:"comment" example:"salary"`
}

type Credit struct {
	UserID  uuid.UUID `json:"user_id" example:"bcf4b5f7-8f73-4205-82e6-cf20e898a98a"`
	Amount  float64   `json:"amount" validate:"gt=0" example:"100"`
	Comment string    `json:"comment" example:"took it from an ATM"`
}
