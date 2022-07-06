package constants

import "time"

const (
	DevEnvironment = "dev"
	Environment    = "Environment"
	CorrelationId  = "X-Correlation-ID"

	JwksUrl      = "JwksUrl"
	JwksAudience = "JwksAudience"
	JwksIssuer   = "JwksIssuer"
	Kid          = "kid"

	JwksResponseKey = "jwksResponse"

	// Cache TTL
	JwksResponseCacheTimeout = 24 * time.Hour

	//Config URL
	ConfigURLPath = "ConfigUrl"

	ServiceNameKey = ""

	//Redis Key
	RedisURLKey               = "redis.url"
	ColonSeparatorForRedisKey = ":"

	// Log file
	LogFile   = "log.file."
	Path      = "path"
	Name      = "name"
	MaxSize   = "maxsize"
	MaxBackUp = "maxbackup"
	MaxAge    = "maxage"

	//Database Keys
	DatabaseKey                   = "database."
	DatabaseUserKey               = "user"
	DatabasePassKey               = "password"
	DatabaseHostKey               = "host"
	DatabaseNameKey               = "name"
	DatabaseMaxIdleConnectionsKey = "maxIdleConnections"
	DatabaseMaxOpenConnectionsKey = "maxOpenConnections"
	DatabaseMaxLifetimeKey        = "maxMaxLifetimeInHours"

	//JWKS
	PEMFilePath = "pkg/docs/key.json"
)
