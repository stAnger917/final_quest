package tests

import (
	"final_quest/configuration"
	"final_quest/internal/repository"
	"final_quest/internal/server"
	"final_quest/internal/usecase/loyality"
	"final_quest/internal/usecase/users"
	"final_quest/pkg/authmw"
	"final_quest/pkg/logging"
	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/runner"
	"log"
	"net/http/httptest"
	"testing"
)

var testDBURI = "user=postgres password=postgres dbname=gophermart_test sslmode=disable"

type TestWCLServer struct {
	cfg *configuration.AppConfig
}

func TestFuncCases(t *testing.T) {
	cfg := configuration.NewConfig()
	err := cfg.InitAppConfiguration()
	if err != nil {
		log.Fatal("failed to init app configuration", err)
	}
	logger := logging.InitAppLogger(cfg.LoggingLevel)
	dbClient, err := repository.NewDBClient(testDBURI)
	if err != nil {
		logger.EasyLogFatal("tests", "failed to init db client on: ", cfg.DatabaseURI, err)
	}
	appRepository := repository.InitAppDB(dbClient, logger)
	err = appRepository.InitTables()
	if err != nil {
		logger.EasyLogFatal("tests", "failed to init db tables", "", err)
	}
	m := mocks.NewNop("loyality")
	err = m.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()

	accrualService := loyality.NewAccountingService(appRepository, logger, m.Service("loyality").ServerAddr())
	appUsersService := users.NewUsersUseCase(appRepository, logger, accrualService)
	appUsersHandler := server.InitAppHandler(appUsersService, logger, accrualService)
	srv := httptest.NewServer(appUsersHandler.Init())
	err = appRepository.PrepareTestData()
	if err != nil {
		logger.EasyLogFatal("tests", "failed to prepare db test data", "", err)
	}

	authmw.PrepareTestTokens()
	defer srv.Close()
	defer dbClient.Close()
	defer appRepository.DropTables()
	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:      srv,
		TestsDir:    "./cases",
		FixturesDir: "fixtures",
		DB:          dbClient,
		Mocks:       m,
	})
}
