package postgresdb

import (
	"fmt"
	"os"
	"time"

	"bitbucket.org/liamstask/goose/lib/goose"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"gitlab.com/feedplan-libraries/common/constants"
	"gitlab.com/feedplan-libraries/common/logger"
)

var db *gorm.DB
var err error

type DBService struct{}

// Init : Initializes the database migrations
func Init() {
	dbUserName := viper.GetString(constants.DatabaseUserKey)
	dbPassword := viper.GetString(constants.DatabasePassKey)
	dbHost := viper.GetString(constants.DatabaseHostKey)
	dbName := viper.GetString(constants.DatabaseNameKey)
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, dbUserName, dbName, dbPassword) //Build connection string
	maxIdleConnections := viper.GetInt(constants.DatabaseMaxIdleConnectionsKey)
	maxOpenConnections := viper.GetInt(constants.DatabaseMaxOpenConnectionsKey)
	connectionMaxLifetime := viper.GetInt(constants.DatabaseMaxLifetimeKey)

	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		fmt.Println("failed to connect.", dbURI, err)
		logger.SugarLogger.Fatalf("Failed to connect to DB", dbURI, err.Error())
		os.Exit(1)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Not able to fetch the working directory")
		logger.SugarLogger.Fatalf("Not able to fetch the working directory")
		os.Exit(1)
	}
	db.DB().SetMaxIdleConns(maxIdleConnections)
	db.DB().SetMaxOpenConns(maxOpenConnections)
	db.DB().SetConnMaxLifetime(time.Hour * time.Duration(connectionMaxLifetime))
	db.SingularTable(true)
	workingDir = workingDir + constants.DatabaseMigrationsScriptPath
	migrateConf := &goose.DBConf{
		MigrationsDir: workingDir,
		Driver: goose.DBDriver{
			Name:    "mysql",
			OpenStr: dbURI,
			Import:  "github.com/go-sql-driver/mysql",
			Dialect: &goose.MySqlDialect{},
		},
	}
	logger.SugarLogger.Infof("Fetching the most recent DB version")
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		logger.SugarLogger.Errorf("Unable to get recent goose db version", err)

	}
	fmt.Println(" Most recent DB version ", latest)
	logger.SugarLogger.Infof("Running the migrations on db", workingDir)
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		logger.SugarLogger.Fatalf("Error while running migrations", err)
		os.Exit(1)
	}
}

// GetDB : Get an instance of DB to connect to the database connection pool
func (d DBService) GetDB() *gorm.DB {
	return db
}
