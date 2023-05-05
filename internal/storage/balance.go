package storage

import (
	"bank/internal/model"
	"bank/pkg/apperror"
	"bank/pkg/postgres"
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:generate moq -out balance_mock.go . BalanceStorage
type BalanceStorage interface {
	WithTransaction(ctx context.Context, fn func(tx postgres.Client) error) error
	Create(ctx context.Context, balance model.Balance) error
	Update(ctx context.Context, balance model.Balance) error
	AtomicUpdate(ctx context.Context, tx postgres.Client, balance model.Balance) error
	Get(ctx context.Context, userID uuid.UUID) (model.Balance, error)
}

type balanceStorage struct {
	client postgres.Client
}

func NewBalanceStorage(client postgres.Client) BalanceStorage {
	return &balanceStorage{client: client}
}

func (storage *balanceStorage) WithTransaction(ctx context.Context, fn func(tx postgres.Client) error) error {
	tx, err := storage.client.Begin(ctx)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				panic(p)
			}
		}
	}()

	err = fn(postgres.Txw{Tx: tx})

	return err
}

func (storage *balanceStorage) Create(ctx context.Context, balance model.Balance) error {
	builder := psql.
		Insert("balance").
		Columns(
			"user_id",
			"balance",
		).
		Values(
			balance.UserID,
			balance.Balance,
		)

	q, args, err := builder.ToSql()
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	_, err = storage.client.Exec(ctx, q, args...)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}

func (storage *balanceStorage) Update(ctx context.Context, balance model.Balance) error {
	return storage.AtomicUpdate(ctx, storage.client, balance)
}

func (storage *balanceStorage) AtomicUpdate(ctx context.Context, tx postgres.Client, balance model.Balance) error {
	builder := psql.
		Update("balance").
		Set("balance", balance.Balance).
		Where(squirrel.Eq{"user_id": balance.UserID})

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

func (storage *balanceStorage) Get(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
	builder := psql.
		Select(
			"user_id",
			"balance",
		).
		From("balance").
		Where(squirrel.Eq{"user_id": userID})

	q, args, err := builder.ToSql()
	if err != nil {
		return model.Balance{}, apperror.Internal.WithError(err)
	}

	var balance model.Balance
	err = storage.client.Get(ctx, &balance, q, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Balance{}, apperror.NotFound.WithError(err)
		}

		return model.Balance{}, apperror.Internal.WithError(err)
	}

	return balance, nil
}
