package main

import (
	"os"
	"transfers-api/internal/config"
	"transfers-api/internal/handlers"
	"transfers-api/internal/logging"
	"transfers-api/internal/repositories"
	"transfers-api/internal/services"
	"transfers-api/internal/transport"
	"transfers-api/internal/version"
)

func main() {
	// init logger
	logger := logging.Logger
	logger.Info("logger started")

	// init config
	cfg := config.ParseFromEnv()
	logger.Infof("config loaded: %v", cfg.String())

	// init repositories
	transfersDB := repositories.NewTransfersMongoDBRepository(cfg.MongoDBConfig)
	// transfersDB := repositories.NewTransfersMySQLRepository(cfg.MySQLConfig)
	transfersCCache := repositories.NewTransfersCCacheRepository(cfg.CCacheConfig)
	logger.Info("repositories created")

	// init services
	transfersService := services.NewTransfersService(cfg.Business, transfersDB, transfersCCache)
	logger.Infof("services created")

	// init handlers
	transfersHandler := handlers.NewTransfersHandler(transfersService)
	logger.Infof("handlers created")

	// init server
	server := transport.NewHTTPServer(transfersHandler)
	server.MapRoutes()
	logger.Infof("server created, running %s@%s", version.AppName, version.Version)

	port := os.Getenv("App_PORT")
	if port == "" {
		port = "8080"
	}

	// run server
	server.Run(":" + port)
}
