package runner

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
	database "github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/handlers"
	"github.com/jim-nnamdi/jinx/pkg/server"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type StartRunner struct {
	ListenAddr string

	LoggingProduction      bool
	LoggingOutputPath      string
	LogggingLevel          string
	ErrorLoggingOutputPath string

	MySQLDatabaseHost     string
	MySQLDatabasePort     string
	MySQLDatabaseUser     string
	MySQLDatabasePassword string
	MySQLDatabaseName     string
}

func (runner *StartRunner) Run(c *cli.Context) error {
	var (
		loggerConfig        = zap.NewDevelopmentConfig()
		logger              *zap.Logger
		err                 error
		mysqlDbInstance     *sql.DB
		mysqlDatabaseClient database.Database
	)
	if runner.LoggingProduction {
		loggerConfig = zap.NewProductionConfig()
		loggerConfig.OutputPaths = []string{runner.LoggingOutputPath}
		loggerConfig.ErrorOutputPaths = []string{runner.ErrorLoggingOutputPath}
	}

	if err = loggerConfig.Level.UnmarshalText([]byte(runner.LogggingLevel)); err != nil {
		return err
	}

	if logger, err = loggerConfig.Build(); err != nil {
		return err
	}

	logger.Sync()
	alog := log.Default()
	databaseConfig := &mysql.Config{
		User:                 runner.MySQLDatabaseUser,
		Passwd:               runner.MySQLDatabasePassword,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", runner.MySQLDatabaseHost, runner.MySQLDatabasePort),
		DBName:               runner.MySQLDatabaseName,
		AllowNativePasswords: true,
	}

	if mysqlDbInstance, err = sql.Open("mysql", databaseConfig.FormatDSN()); err != nil {
		return fmt.Errorf("unable to open connection to MySQL Server: %s", err.Error())
	}

	if mysqlDatabaseClient, err = database.NewMySQLDatabase(mysqlDbInstance); err != nil {
		return fmt.Errorf("unable to create MySQL database client: %s", err.Error())
	}
	server := &server.GracefulShutdownServer{
		HTTPListenAddr:     runner.ListenAddr,
		RegisterHandler:    handlers.NewRegisterHandler(logger, mysqlDatabaseClient),
		LoginHandler:       handlers.NewLoginHandler(logger, mysqlDatabaseClient),
		ProfileHandler:     handlers.NewProfileHandler(logger, mysqlDatabaseClient),
		HomeHandler:        handlers.NewHomeHandler(),
		AddForumHandler:    handlers.NewForumStruct(alog, mysqlDatabaseClient),
		AllForumHandler:    handlers.NewAForumStruct(alog, mysqlDatabaseClient),
		SingleForumHandler: handlers.NewSForumStruct(alog, mysqlDatabaseClient),
		ChatHandler:        handlers.NewChat(alog, mysqlDatabaseClient),
	}
	server.Start()
	return nil
}
