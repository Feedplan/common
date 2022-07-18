package postgresdb

import (
	"fmt"
	"os"
	"time"

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
	db.DB().SetMaxIdleConns(maxIdleConnections)
	db.DB().SetMaxOpenConns(maxOpenConnections)
	db.DB().SetConnMaxLifetime(time.Hour * time.Duration(connectionMaxLifetime))
	db.SingularTable(true)
}

// GetDB : Get an instance of DB to connect to the database connection pool
func (d DBService) GetDB() *gorm.DB {
	return db
}
