package handler

import (
	"bank/internal/domain"
	"bank/internal/dto"
	"bank/internal/service"
	"bank/pkg/apilayer"
	"bank/pkg/apperror"
	"bank/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

type BalanceHandler struct {
	balanceService service.BalanceService
	apiLayer       *apilayer.Client
}

func NewBalanceHandler(balanceService service.BalanceService, apilayer *apilayer.Client) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService, apiLayer: apilayer}
}

func (handler *BalanceHandler) Register(router fiber.Router) {
	router.Get("", handler.Get)
	router.Get("/transaction", handler.SelectTransaction)
	router.Post("/transfer", handler.Transfer)
	router.Post("/debet", handler.Debet)
	router.Post("/credit", handler.Credit)
}

// Get godoc
// @Summary Retrieves balance based on given user ID
// @Produce json
// @Param	user_id  	query 	string 	true	"user id" 	Format(uuid)
// @Param	currency	query 	string 	false	"currency"	Format(string)
// @Success 200 {object} domain.Balance
// @Failure 404 {object} apperror.Error
// @Router /api/v1/balance [get]
func (handler *BalanceHandler) Get(c *fiber.Ctx) error {
	var request dto.GetBalance
	if err := c.QueryParser(&request); err != nil {
		return apperror.BadRequest.WithError(err)
	}

	err := validator.Validate(request)
	if err != nil {
		return err
	}

	var balance domain.Balance
	balance, err = handler.balanceService.Get(c.Context(), request.UserID)
	if err != nil {
		return err
	}

	if request.Currency != "" {
		balance.Balance, err = handler.apiLayer.Convert(balance.Balance, "RUB", request.Currency)
		if err != nil {
			return apperror.BadRequest.WithError(err).WithMessage("invalid currency")
		}
	}

	return c.JSON(fiber.Map{
		"response": balance,
	})
}

// SelectTransaction godoc
// @Summary Retrieves transactions based on given user ID
// @Produce json
// @Param	user_id  	query 	string 	true	"user id" 	Format(uuid)
// @Param	sort	    query 	string 	false	"sort"		Enums(amount, created_at)
// @Param	order	    query 	string 	false	"order"		Enums(asc, desc)
// @Param	count	    query 	int 	false	"count"		Format(int)
// @Param	offset	    query 	int 	false	"offset"	Format(int)
// @Success 200 {array} domain.Transaction
// @Failure 404 {object} apperror.Error
// @Router /api/v1/balance/transaction [get]
func (handler *BalanceHandler) SelectTransaction(c *fiber.Ctx) error {
	var request dto.SelectTransaction
	if err := c.QueryParser(&request); err != nil {
		return apperror.BadRequest.WithError(err)
	}

	err := validator.Validate(request)
	if err != nil {
		return err
	}

	var transactions []domain.Transaction
	transactions, err = handler.balanceService.SelectTransaction(c.Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"response": transactions,
	})
}

// Transfer godoc
// @Summary Transfer money from payer id to payee id
// @Produce json
// @Param request body dto.Transfer true "transfer params"
// @Success 200 {int} 1
// @Failure 404 {object} apperror.Error
// @Router /api/v1/balance/transfer [post]
func (handler *BalanceHandler) Transfer(c *fiber.Ctx) error {
	var request dto.Transfer
	if err := c.BodyParser(&request); err != nil {
		return apperror.BadRequest.WithError(err)
	}

	err := validator.Validate(request)
	if err != nil {
		return err
	}

	err = handler.balanceService.Transfer(c.Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"response": 1,
	})
}

// Debet godoc
// @Summary Deposits money to the balance by user id
// @Produce json
// @Param request body dto.Debet true "debet params"
// @Success 200 {object} domain.Balance
// @Failure 404 {object} apperror.Error
// @Router /api/v1/balance/debet [post]
func (handler *BalanceHandler) Debet(c *fiber.Ctx) error {
	var request dto.Debet
	if err := c.BodyParser(&request); err != nil {
		return apperror.BadRequest.WithError(err)
	}

	err := validator.Validate(request)
	if err != nil {
		return err
	}

	var balance domain.Balance
	balance, err = handler.balanceService.Debet(c.Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"response": balance,
	})
}

// Credit godoc
// @Summary Withdraws money from the balance by user id
// @Produce json
// @Param request body dto.Credit true "credit params"
// @Success 200 {object} domain.Transaction
// @Failure 404 {object} apperror.Error
// @Router /api/v1/balance/credit [post]
func (handler *BalanceHandler) Credit(c *fiber.Ctx) error {
	var request dto.Credit
	if err := c.BodyParser(&request); err != nil {
		return apperror.BadRequest.WithError(err)
	}

	err := validator.Validate(request)
	if err != nil {
		return err
	}

	var balance domain.Balance
	balance, err = handler.balanceService.Credit(c.Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"response": balance,
	})
}
