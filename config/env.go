package config

import (
	"os"

	"gitlab.com/feedplan-libraries/common/constants"
)

type Env string

const (
	Dev  Env = constants.DevEnvironment
	Prod Env = constants.ProdEnvironment
)

func InDev() bool {
	return GetEnv() == Dev
}

func InProd() bool {
	return GetEnv() == Prod
}

func GetEnv() Env {
	environment := os.Getenv("BOOT_CUR_ENV")
	switch environment {
	case string(Dev):
		return Dev
	case string(Prod):
		return Prod
	default:
		return Dev
	}
}