package service_test

import (
	"bank/internal/domain"
	"bank/internal/dto"
	"bank/internal/model"
	"bank/internal/service"
	"bank/internal/storage"
	"bank/pkg/apperror"
	"bank/pkg/postgres"
	"bank/pkg/sort"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func DeepEqualWithZero(obj1, obj2 interface{}) bool {
	if reflect.DeepEqual(obj1, obj2) {
		return true
	}

	value1 := reflect.ValueOf(obj1)
	value2 := reflect.ValueOf(obj2)

	if value1.Kind() == reflect.Ptr {
		value1 = value1.Elem()
	}
	if value2.Kind() == reflect.Ptr {
		value2 = value2.Elem()
	}

	if value1.Type() != value2.Type() {
		return false
	}

	for i := 0; i < value1.NumField(); i++ {
		if !reflect.DeepEqual(value1.Field(i).Interface(), value2.Field(i).Interface()) && !reflect.ValueOf(value2.Field(i).Interface()).IsZero() {
			return false
		}
	}

	return true
}

func TestBalanceService_Get(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		balanceStorage     *storage.BalanceStorageMock
		transactionStorage *storage.TransactionStorageMock
		request            dto.GetBalance
		response           domain.Balance
		wantErr            error
	}{
		{
			name: "ok",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 10000}, nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{},
			request:            dto.GetBalance{UserID: userID, Currency: "RUB"},
			response:           domain.Balance{UserID: userID, Balance: 100.0},
			wantErr:            nil,
		},
		{
			name: "balance not found",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{}, apperror.NotFound
				},
			},
			transactionStorage: &storage.TransactionStorageMock{},
			request:            dto.GetBalance{UserID: userID, Currency: "RUB"},
			response:           domain.Balance{},
			wantErr:            apperror.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceService := service.NewBalanceService(tt.balanceStorage, tt.transactionStorage)

			got, err := balanceService.Get(context.Background(), tt.request.UserID)
			if !DeepEqualWithZero(got, tt.response) {
				t.Errorf("BalanceService.Get() = %v, want %v", got, tt.response)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("BalanceService.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBalanceService_Transfer(t *testing.T) {
	payeeID := uuid.New()
	payerID := uuid.New()

	tests := []struct {
		name               string
		balanceStorage     *storage.BalanceStorageMock
		transactionStorage *storage.TransactionStorageMock
		request            dto.Transfer
		response           error
	}{
		{
			name: "ok",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 10000}, nil
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Transfer{
				PayeeID: payeeID,
				PayerID: payerID,
				Amount:  100,
				Comment: "comment",
			},
			response: nil,
		},
		{
			name: "insufficient funds",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 0}, nil
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return apperror.BadRequest
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Transfer{
				PayeeID: payeeID,
				PayerID: payerID,
				Amount:  10000,
				Comment: "comment",
			},
			response: apperror.BadRequest,
		},
		{
			name: "balance not found",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{}, apperror.NotFound
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Transfer{
				PayeeID: payeeID,
				PayerID: payerID,
				Amount:  10000,
				Comment: "comment",
			},
			response: apperror.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceService := service.NewBalanceService(tt.balanceStorage, tt.transactionStorage)

			err := balanceService.Transfer(context.Background(), tt.request)
			if !errors.Is(err, tt.response) {
				t.Errorf("BalanceService.Transfer() error = %v, wantErr %v", err, tt.response)
			}
		})
	}
}

func TestBalanceService_Debet(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		balanceStorage     *storage.BalanceStorageMock
		transactionStorage *storage.TransactionStorageMock
		request            dto.Debet
		response           domain.Balance
	}{
		{
			name: "ok",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 0}, nil
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Debet{
				UserID: userID,
				Amount: 100.0,
			},
			response: domain.Balance{
				UserID:  userID,
				Balance: 100.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceService := service.NewBalanceService(tt.balanceStorage, tt.transactionStorage)

			transaction, err := balanceService.Debet(context.Background(), tt.request)
			if !DeepEqualWithZero(transaction, tt.response) {
				t.Errorf("BalanceService.Debet() = %v, want %v", transaction, tt.response)
			}

			if !errors.Is(err, nil) {
				t.Errorf("BalanceService.Debet() error = %v, wantErr %v", err, nil)
			}
		})
	}
}

func TestBalanceService_Credit(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		balanceStorage     *storage.BalanceStorageMock
		transactionStorage *storage.TransactionStorageMock
		request            dto.Credit
		response           domain.Balance
		wantErr            error
	}{
		{
			name: "insufficient funds",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 50}, nil
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Credit{
				UserID: userID,
				Amount: 100.0,
			},
			response: domain.Balance{},
			wantErr:  apperror.BadRequest,
		},
		{
			name: "ok",
			balanceStorage: &storage.BalanceStorageMock{
				GetFunc: func(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
					return model.Balance{UserID: userID, Balance: 10000}, nil
				},
				UpdateFunc: func(ctx context.Context, balance model.Balance) error {
					return nil
				},
				WithTransactionFunc: func(ctx context.Context, fn func(tx postgres.Client) error) error {
					return nil
				},
			},
			transactionStorage: &storage.TransactionStorageMock{
				CreateFunc: func(ctx context.Context, transaction model.Transaction) error {
					return nil
				},
			},
			request: dto.Credit{
				UserID: userID,
				Amount: 10.0,
			},
			response: domain.Balance{
				UserID:  userID,
				Balance: 90.0,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceService := service.NewBalanceService(tt.balanceStorage, tt.transactionStorage)

			transaction, err := balanceService.Credit(context.Background(), tt.request)
			if err != nil && tt.wantErr == nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantErr != nil && errors.Is(err, tt.wantErr) {
				return
			}

			if !DeepEqualWithZero(transaction, tt.response) {
				t.Errorf("BalanceService.Credit() = %v, want %v", transaction, tt.response)
			}
		})
	}
}

func TestBalanceService_SelectTransaction(t *testing.T) {
	balanceStorage := &storage.BalanceStorageMock{}
	transactionStorage := &storage.TransactionStorageMock{
		SelectFunc: func(ctx context.Context, userID uuid.UUID, filter sort.Filter) ([]model.Transaction, error) {
			return []model.Transaction{
				{
					PayeeID: userID,
					Type:    domain.Debet,
					Amount:  100.0,
				},
				{
					PayeeID: userID,
					Type:    domain.Credit,
					Amount:  -100.0,
				},
			}, nil
		},
	}

	balanceService := service.NewBalanceService(balanceStorage, transactionStorage)

	userID := uuid.New()
	transactions, err := balanceService.SelectTransaction(context.Background(), dto.SelectTransaction{
		UserID: userID,
	})
	assert.NoError(t, err)

	assert.Equal(t, 2, len(transactions))
}
