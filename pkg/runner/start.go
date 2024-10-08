package runner

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	database "github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/handlers"
	"github.com/jim-nnamdi/jinx/pkg/server"
	"github.com/jim-nnamdi/jinx/pkg/utils"
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
	databaseConfig := &mysql.Config{
		User:                 runner.MySQLDatabaseUser,
		Passwd:               runner.MySQLDatabasePassword,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", runner.MySQLDatabaseHost, runner.MySQLDatabasePort),
		DBName:               runner.MySQLDatabaseName,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if mysqlDbInstance, err = sql.Open("mysql", databaseConfig.FormatDSN()); err == nil {
			if err = mysqlDbInstance.Ping(); err == nil {
				// Successfully connected
				break
			}
		}
		log.Printf("Failed to connect to MySQL database, attempt %d: %v", i+1, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		utils.Logger.Info("err connecting to database after multiple tries")
		return fmt.Errorf("unable to open connection to MySQL Server: %s", err.Error())
	}

	if mysqlDatabaseClient, err = database.NewMySQLDatabase(mysqlDbInstance); err != nil {
		return fmt.Errorf("unable to create MySQL database client: %s", err.Error())
	}
	utils.Logger.Info("connected to database successfully")
	server := &server.GracefulShutdownServer{
		HTTPListenAddr:     runner.ListenAddr,
		RegisterHandler:    handlers.NewRegisterHandler(logger, mysqlDatabaseClient),
		LoginHandler:       handlers.NewLoginHandler(logger, mysqlDatabaseClient),
		ProfileHandler:     handlers.NewProfileHandler(logger, mysqlDatabaseClient),
		HomeHandler:        handlers.NewHomeHandler(),
		AddForumHandler:    handlers.NewForumStruct(logger, mysqlDatabaseClient),
		AllForumHandler:    handlers.NewAForumStruct(logger, mysqlDatabaseClient),
		SingleForumHandler: handlers.NewSForumStruct(logger, mysqlDatabaseClient),
		ChatHandler:        handlers.NewChat(logger, mysqlDatabaseClient),
		CommentHandler:     handlers.NewCommentHandler(logger, mysqlDatabaseClient),
		CreateGroup:        handlers.NewCreateGroupHandler(logger, mysqlDatabaseClient),
		AddUserToGroup:     handlers.NewAddGroupMemberHandler(logger, mysqlDatabaseClient),
		SendGroupMessage:   handlers.NewSendGroupMessageHandler(logger, mysqlDatabaseClient),
		GetChatHistory:     handlers.NewGetUserChatsHistoryHandler(logger, mysqlDatabaseClient),
	}
	server.Start()
	return nil
}
