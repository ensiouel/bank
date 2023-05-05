package app

import (
	"bank/internal/config"
	"bank/internal/service"
	"bank/internal/storage"
	"bank/internal/transport"
	"bank/internal/transport/handler"
	"bank/pkg/apilayer"
	"bank/pkg/postgres"
	"context"
	"errors"
	"github.com/pressly/goose"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

type App struct {
	config config.Config
}

func New() *App {
	conf, err := config.New()
	if err != nil {
		log.Panic(err)
	}

	return &App{
		config: conf,
	}
}

func (app *App) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgConfig := postgres.Config{
		Host: app.config.Postgres.Host, Port: app.config.Postgres.Port, DB: app.config.Postgres.DB,
		User: app.config.Postgres.User, Password: app.config.Postgres.Password,
	}
	pgClient, err := postgres.NewClient(ctx, pgConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err = migrate("up", "migration", pgConfig.String()); err != nil {
		log.Fatal(err)
	}

	apiLayerClient := apilayer.New(app.config.APILayer.APIKey)

	transactionStorage := storage.NewTransactionStorage(pgClient)

	balanceStorage := storage.NewBalanceStorage(pgClient)
	balanceService := service.NewBalanceService(balanceStorage, transactionStorage)
	balanceHandler := handler.NewBalanceHandler(balanceService, apiLayerClient)

	go func() {
		err = transport.New().Handle(balanceHandler).Listen(app.config.Server.Addr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
}

func migrate(command string, dir string, dbstring string) error {
	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = goose.Run(command, db, dir); err != nil {
		return err
	}

	return nil
}
