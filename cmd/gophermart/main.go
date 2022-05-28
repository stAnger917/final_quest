package main

import (
	"context"
	"final_quest/configuration"
	"final_quest/internal/repository"
	"final_quest/internal/server"
	"final_quest/internal/usecase/users"
	"final_quest/pkg/logging"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := configuration.NewConfig()
	err := cfg.InitAppConfiguration()
	if err != nil {
		log.Fatal("failed to init app configuration", err)
	}
	logger := logging.InitAppLogger(cfg.LoggingLevel)
	dbClient, err := repository.NewDBClient(cfg.DatabaseURI)
	if err != nil {
		logger.EasyLogFatal("main", "failed to init db client on: ", cfg.DatabaseURI, err)
	}

	appRepository := repository.InitAppDB(dbClient, logger)
	err = appRepository.InitTables()
	if err != nil {
		logger.EasyLogFatal("main", "failed to init db tables", "", err)
	}
	appUsersService := users.NewUsersUseCase(appRepository, logger)
	appUsersHandler := server.InitAppHandler(appUsersService, logger)
	srv := server.InitNewServer(cfg.RunAddress, appUsersHandler.Init())
	err = srv.Run()
	logger.EasyLogInfo("main", "running server on: ", cfg.RunAddress)
	if err != nil {
		logger.EasyLogFatal("main", "failed to run server on: ", cfg.RunAddress, err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.EasyLogError("main", "failed to safe shutdown server", "", err)
	}

	if err := dbClient.Close(); err != nil {
		logger.EasyLogError("main", "failed to safe shutdown db", "", err)
	}
}
