package constants

import "time"

const (
	DevEnvironment = "dev"
	Environment    = "Environment"
	CorrelationId  = "X-Correlation-ID"
	ServiceNameKey = "serviceName"
	AwsRegionKey   = "awsRegion"
	ConfigFilePath = "/pkg/config/config.json"
	ServerPortKey  = "serverPort"

	//Jwks
	JwksAudience    = "JwksAudience"
	JwksIssuer      = "JwksIssuer"
	Kid             = "kid"
	JwksResponseKey = "jwksResponse"

	// Cache TTL
	JwksResponseCacheTimeout = 24 * time.Hour

	//Redis Key
	RedisURLKey               = "redisUrl"
	ColonSeparatorForRedisKey = ":"

	// Log file
	LogFilePath      = "logFilePath"
	LogFileName      = "logName"
	LogFileMaxSize   = "logMaxSize"
	LogFileMaxBackUp = "logMaxBackUp"
	LogFileMaxAge    = "logMaxAge"

	//Database Keys
	DatabaseUserKey               = "dbUser"
	DatabasePassKey               = "dbPassword"
	DatabaseHostKey               = "dbHost"
	DatabaseNameKey               = "dbName"
	DatabaseMaxIdleConnectionsKey = "dbMaxIdleConnections"
	DatabaseMaxOpenConnectionsKey = "dbMaxOpenConnections"
	DatabaseMaxLifetimeKey        = "dbMaxMaxLifetimeInHours"
	DatabaseMigrationsScriptPath  = "/internal/app/db/migrations"
)
