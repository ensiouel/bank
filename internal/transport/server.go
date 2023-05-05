package transport

import (
	_ "bank/docs"
	"bank/internal/transport/handler"
	"bank/internal/transport/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

type Server struct {
	router *fiber.App
}

func New() *Server {
	router := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(),
	})

	router.Use(
		recover.New(),
		logger.New(),
	)

	return &Server{
		router: router,
	}
}

func (server *Server) Handle(balanceHandler *handler.BalanceHandler) *Server {
	server.router.Group("/swagger/*", fiberSwagger.WrapHandler)

	api := server.router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			balanceHandler.Register(v1.Group("/balance"))
		}
	}

	return server
}

func (server *Server) Listen(addr string) error {
	return server.router.Listen(addr)
}
