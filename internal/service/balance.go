package service

import (
	"bank/internal/domain"
	"bank/internal/dto"
	"bank/internal/model"
	"bank/internal/storage"
	"bank/pkg/apperror"
	"bank/pkg/postgres"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type BalanceService interface {
	Get(ctx context.Context, userID uuid.UUID) (domain.Balance, error)
	Transfer(ctx context.Context, request dto.Transfer) error
	Debet(ctx context.Context, request dto.Debet) (domain.Balance, error)
	Credit(ctx context.Context, request dto.Credit) (domain.Balance, error)
	SelectTransaction(ctx context.Context, request dto.SelectTransaction) ([]domain.Transaction, error)
}

type balanceService struct {
	balanceStorage     storage.BalanceStorage
	transactionStorage storage.TransactionStorage
}

func NewBalanceService(
	balanceStorage storage.BalanceStorage,
	transactionStorage storage.TransactionStorage,
) BalanceService {
	return &balanceService{
		balanceStorage:     balanceStorage,
		transactionStorage: transactionStorage,
	}
}

func (service *balanceService) Get(ctx context.Context, userID uuid.UUID) (domain.Balance, error) {
	balance, err := service.get(ctx, userID, false)
	if err != nil {
		return domain.Balance{}, err
	}

	return domain.BalanceFromModel(balance), nil
}

func (service *balanceService) Transfer(ctx context.Context, request dto.Transfer) error {
	if request.PayerID == request.PayeeID {
		return apperror.BadRequest.WithMessage("transfer is not possible with yourself")
	}

	payerBalance, err := service.get(ctx, request.PayerID, false)
	if err != nil {
		return err
	}

	payeeBalance, err := service.get(ctx, request.PayeeID, false)
	if err != nil {
		return err
	}

	err = service.balanceStorage.WithTransaction(ctx, func(tx postgres.Client) error {
		payerBalance.Balance -= int64(request.Amount * 100.0)

		if payerBalance.Balance < 0 {
			return apperror.BadRequest.WithMessage("insufficient funds")
		}

		err = service.balanceStorage.AtomicUpdate(ctx, tx, payerBalance)
		if err != nil {
			if apperr, ok := apperror.Is(err, apperror.Internal); ok {
				return apperr.WithScope("balanceService.Transfer")
			}

			return err
		}

		payeeBalance.Balance += int64(request.Amount * 100.0)

		err = service.balanceStorage.AtomicUpdate(ctx, tx, payeeBalance)
		if err != nil {
			if apperr, ok := apperror.Is(err, apperror.Internal); ok {
				return apperr.WithScope("balanceService.Transfer")
			}

			return err
		}

		now := time.Now()

		_, err = service.atomicRecord(ctx, tx, domain.Transfer, request.PayeeID, &request.PayerID, int64(request.Amount*100.0), request.Comment, now)
		if err != nil {
			return err
		}

		_, err = service.atomicRecord(ctx, tx, domain.Transfer, request.PayerID, &request.PayeeID, -int64(request.Amount*100.0), request.Comment, now)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.Internal); ok {
			return apperr.WithScope("balanceService.WithTransaction")
		}

		return err
	}

	return nil
}

func (service *balanceService) Debet(ctx context.Context, request dto.Debet) (domain.Balance, error) {
	balance, err := service.get(ctx, request.UserID, true)
	if err != nil {
		return domain.Balance{}, err
	}

	balance.Balance += int64(request.Amount * 100.0)

	err = service.balanceStorage.WithTransaction(ctx, func(tx postgres.Client) error {
		err = service.balanceStorage.AtomicUpdate(ctx, tx, balance)
		if err != nil {
			if apperr, ok := apperror.Is(err, apperror.Internal); ok {
				return apperr.WithScope("balanceService.Debet")
			}

			return err
		}

		now := time.Now()

		_, err = service.atomicRecord(ctx, tx, domain.Debet, request.UserID, nil, int64(request.Amount*100.0), request.Comment, now)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.Internal); ok {
			return domain.Balance{}, apperr.WithScope("balanceService.WithTransaction")
		}

		return domain.Balance{}, err
	}

	return domain.BalanceFromModel(balance), nil
}

func (service *balanceService) Credit(ctx context.Context, request dto.Credit) (domain.Balance, error) {
	balance, err := service.get(ctx, request.UserID, false)
	if err != nil {
		return domain.Balance{}, err
	}

	balance.Balance -= int64(request.Amount * 100.0)

	if balance.Balance < 0 {
		return domain.Balance{}, apperror.BadRequest.WithMessage("insufficient funds")
	}

	err = service.balanceStorage.WithTransaction(ctx, func(tx postgres.Client) error {
		err = service.balanceStorage.AtomicUpdate(ctx, tx, balance)
		if err != nil {
			if apperr, ok := apperror.Is(err, apperror.Internal); ok {
				return apperr.WithScope("balanceService.Credit")
			}

			return err
		}

		now := time.Now()

		_, err = service.atomicRecord(ctx, tx, domain.Credit, request.UserID, nil, -int64(request.Amount*100.0), request.Comment, now)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.Internal); ok {
			return domain.Balance{}, apperr.WithScope("balanceService.WithTransaction")
		}

		return domain.Balance{}, err
	}

	return domain.BalanceFromModel(balance), nil
}

func (service *balanceService) SelectTransaction(ctx context.Context, request dto.SelectTransaction) ([]domain.Transaction, error) {
	transactions, err := service.transactionStorage.Select(ctx, request.UserID, request.SelectTransactionFilter)
	if err != nil {
		if !errors.Is(err, apperror.NotFound) {
			if apperr, ok := apperror.Is(err, apperror.Internal); ok {
				return []domain.Transaction{}, apperr.WithScope("balanceService.SelectTransaction")
			}

			return []domain.Transaction{}, err
		}
	}

	return domain.TransactionsFromModels(transactions), nil
}

func (service *balanceService) get(ctx context.Context, userID uuid.UUID, forceCreate bool) (model.Balance, error) {
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	balance, err := service.balanceStorage.Get(ctx, userID)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.NotFound); ok {
			if !forceCreate {
				return balance, apperr.WithMessage("balance not found")
			}

			balance = model.Balance{
				UserID:  userID,
				Balance: 0,
			}

			err = service.balanceStorage.Create(ctx, balance)
			if err != nil {
				if apperr, ok = apperror.Is(err, apperror.Internal); ok {
					return balance, apperr.WithScope("balanceService.get")
				}

				return balance, err
			}
		} else {
			if apperr, ok = apperror.Is(err, apperror.Internal); ok {
				return balance, apperr.WithScope("balanceService.get")
			}

			return balance, err
		}
	}

	return balance, nil
}

func (service *balanceService) record(ctx context.Context, transactionType string, payeeID uuid.UUID, payerID *uuid.UUID, amount int64, comment string, createdAt time.Time) (model.Transaction, error) {
	transaction := model.Transaction{
		ID:        uuid.New(),
		PayeeID:   payeeID,
		PayerID:   payerID,
		Type:      transactionType,
		Amount:    amount,
		Comment:   comment,
		CreatedAt: createdAt,
	}
	err := service.transactionStorage.Create(ctx, transaction)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.Internal); ok {
			return transaction, apperr.WithScope("balanceService.record")
		}

		return transaction, err
	}

	return transaction, nil
}

func (service *balanceService) atomicRecord(ctx context.Context, tx postgres.Client, transactionType string, payeeID uuid.UUID, payerID *uuid.UUID, amount int64, comment string, createdAt time.Time) (model.Transaction, error) {
	transaction := model.Transaction{
		ID:        uuid.New(),
		PayeeID:   payeeID,
		PayerID:   payerID,
		Type:      transactionType,
		Amount:    amount,
		Comment:   comment,
		CreatedAt: createdAt,
	}
	err := service.transactionStorage.AtomicCreate(ctx, tx, transaction)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.Internal); ok {
			return transaction, apperr.WithScope("balanceService.record")
		}

		return transaction, err
	}

	return transaction, nil
}
