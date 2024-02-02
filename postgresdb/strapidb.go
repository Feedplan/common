package postgresdb

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"gitlab.com/feedplan-libraries/common/constants"
	"gitlab.com/feedplan-libraries/common/logger"
)

var strapiDB *gorm.DB
var strapiErr error

// InitializeStrapiDB : Initializes the database migrations
func InitializeStrapiDB() {
	dbUserName := viper.GetString(constants.DatabaseUserKey)
	dbPassword := viper.GetString(constants.DatabasePassKey)
	dbHost := viper.GetString(constants.DatabaseHostKey)
	dbName := viper.GetString(constants.StrapiDBKey)
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, dbUserName, dbName, dbPassword) //Build connection string
	maxIdleConnections := viper.GetInt(constants.DatabaseMaxIdleConnectionsKey)
	maxOpenConnections := viper.GetInt(constants.DatabaseMaxOpenConnectionsKey)
	connectionMaxLifetime := viper.GetInt(constants.DatabaseMaxLifetimeKey)

	strapiDB, strapiErr = gorm.Open("postgres", dbURI)
	if strapiErr != nil {
		fmt.Println("failed to connect.", dbURI, strapiErr)
		logger.SugarLogger.Fatalf("Failed to connect to feedplan-strapi DB", dbURI, strapiErr.Error())
		os.Exit(1)
	}
	strapiDB.DB().SetMaxIdleConns(maxIdleConnections)
	strapiDB.DB().SetMaxOpenConns(maxOpenConnections)
	strapiDB.DB().SetConnMaxLifetime(time.Hour * time.Duration(connectionMaxLifetime))
	strapiDB.SingularTable(true)
}

// GetStrapiDB : Get an instance of DB to connect to the database connection pool
func (d DBService) GetStrapiDB() *gorm.DB {
	return strapiDB
}
