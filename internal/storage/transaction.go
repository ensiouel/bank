package storage

import (
	"bank/internal/model"
	"bank/pkg/apperror"
	"bank/pkg/postgres"
	"bank/pkg/sort"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:generate moq -out transaction_mock.go . TransactionStorage
type TransactionStorage interface {
	Create(ctx context.Context, transaction model.Transaction) error
	AtomicCreate(ctx context.Context, tx postgres.Client, transaction model.Transaction) error
	Select(ctx context.Context, userID uuid.UUID, filter sort.Filter) ([]model.Transaction, error)
}

type transactionStorage struct {
	client postgres.Client
}

func NewTransactionStorage(client postgres.Client) TransactionStorage {
	return &transactionStorage{client: client}
}

func (storage *transactionStorage) Create(ctx context.Context, transaction model.Transaction) error {
	return storage.AtomicCreate(ctx, storage.client, transaction)
}

func (storage *transactionStorage) AtomicCreate(ctx context.Context, tx postgres.Client, transaction model.Transaction) error {
	builder := psql.
		Insert("transaction").
		Columns(
			"id",
			"payee_id",
			"payer_id",
			"type",
			"amount",
			"comment",
			"created_at",
		).
		Values(
			transaction.ID,
			transaction.PayeeID,
			transaction.PayerID,
			transaction.Type,
			transaction.Amount,
			transaction.Comment,
			transaction.CreatedAt,
		)

	q, args, err := builder.ToSql()
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	_, err = tx.Exec(ctx, q, args...)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}

func (storage *transactionStorage) Select(ctx context.Context, userID uuid.UUID, filter sort.Filter) ([]model.Transaction, error) {
	builder := psql.
		Select(
			"id",
			"payee_id",
			"payer_id",
			"type",
			"amount",
			"comment",
			"created_at",
		).
		From("transaction").
		Where(squirrel.Eq{"payee_id": userID}).
		Offset(filter.GetOffset())

	if filter.GetCount() != 0 {
		builder = builder.
			Limit(filter.GetCount())
	}

	if filter.GetSort() != "" {
		builder = builder.
			OrderBy(fmt.Sprintf("%s %s", filter.GetSort(), filter.GetOrder()))
	}

	q, args, err := builder.ToSql()
	if err != nil {
		return []model.Transaction{}, apperror.Internal.WithError(err)
	}

	var transactions []model.Transaction
	err = storage.client.Select(ctx, &transactions, q, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []model.Transaction{}, apperror.NotFound.WithError(err)
		}

		return []model.Transaction{}, err
	}

	return transactions, nil
}
