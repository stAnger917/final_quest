package tests

import (
	"final_quest/configuration"
	"final_quest/internal/repository"
	"final_quest/internal/server"
	"final_quest/internal/usecase/users"
	"final_quest/pkg/authMW"
	"final_quest/pkg/logging"
	"github.com/lamoda/gonkey/runner"
	"log"
	"net/http/httptest"
	"testing"
)

var testDBURI = "user=postgres password=postgres dbname=gophermart_test sslmode=disable"

func TestFuncCases(t *testing.T) {
	cfg := configuration.NewConfig()
	err := cfg.InitAppConfiguration()
	if err != nil {
		log.Fatal("failed to init app configuration", err)
	}
	logger := logging.InitAppLogger(cfg.LoggingLevel)
	dbClient, err := repository.NewDBClient(testDBURI)
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
	srv := httptest.NewServer(appUsersHandler.Init())
	authMW.PrepareTestTokens()
	defer srv.Close()
	defer dbClient.Close()
	defer appRepository.DropTables()
	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:      srv,
		TestsDir:    "./cases",
		FixturesDir: "fixtures",
		DB:          dbClient,
	})
}
