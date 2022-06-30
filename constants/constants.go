package constants

import "time"

const (
	DevEnvironment = "dev"
	Environment    = "Environment"

	JwksUrl      = "JwksUrl"
	JwksAudience = "JwksAudience"
	JwksIssuer   = "JwksIssuer"
	Kid          = "kid"

	ColonSeparator  = ":"
	JwksResponseKey = "jwksResponse"

	// Cache TTL
	JwksResponseCacheTimeout = 24 * time.Hour

	//Config URL
	ConfigURLPath = "ConfigUrl"
)
