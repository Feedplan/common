package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/feedplan/common/constants"
	"github.com/feedplan/common/logger"
	"github.com/spf13/viper"
)

// Init :
func Init(service, env, region string) {
	addSysConfig()
	body, err := fetchConfiguration(service, env, region)
	if err != nil {
		fmt.Println("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
	}
	parseConfiguration(body, env)
}

// Make HTTP request to fetch configuration from config server
func fetchConfiguration(service, env, region string) ([]byte, error) {
	var bodyBytes []byte
	var err error
	inDev := strings.Compare(env, constants.DevEnvironment) == 0
	inLocal := strings.Compare(env, constants.LocalEnvironment) == 0
	inProd := strings.Compare(env, constants.ProdEnvironment) == 0
	if inDev || inProd || inLocal {
		//panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Println("Not able to fetch the working directory")
			logger.SugarLogger.Fatalf("Not able to fetch the working directory")
			os.Exit(1)
		}
		bodyBytes, err = ioutil.ReadFile(workingDir + constants.ConfigFilePath)
		if err != nil {
			fmt.Println("Couldn't read local configuration file.", err)
		} else {
			log.Print("using local config.")
		}
	} else {
		url := "https://feedplan-" + env + ".s3." + region + ".amazonaws.com/" + service + "/config.json"
		fmt.Printf("url is : %s \n", url)
		fmt.Printf("Loading config from %s \n", url)
		resp, err := http.Get(url)
		if resp != nil || err == nil {
			bodyBytes, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading configuration response body.")
				logger.SugarLogger.Fatalf("Error reading configuration response body.")
			}
		}
	}
	return bodyBytes, err
}

// Get DB and cred from sys env
func addSysConfig() {
	dbUser := getEnvOrDefault("DB_USERNAME", "postgres")
	viper.Set(constants.DatabaseUserKey, dbUser)
	dbPassord := getEnvOrDefault("DB_PASSWORD", "flywaydb")
	viper.Set(constants.DatabasePassKey, dbPassord)
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	viper.Set(constants.DatabaseHostKey, dbHost)
}

func getEnvOrDefault(envKey, defaultValue string) string {
	envValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	}
	return envValue
}

// Pass JSON bytes into struct and then into Viper
func parseConfiguration(body []byte, env string) {
	var cloudConfig springCloudConfig
	err := json.Unmarshal(body, &cloudConfig)
	if err != nil {
		fmt.Println("Cannot parse configuration, message: " + err.Error())
	}
	var sources map[string]interface{}
	inDev := strings.Compare(env, constants.DevEnvironment) == 0
	inLocal := strings.Compare(env, constants.LocalEnvironment) == 0
	inProd := strings.Compare(env, constants.ProdEnvironment) == 0
	if inDev {
		sources = cloudConfig.PropertySources.Source
	}
	if inProd {
		sources = cloudConfig.PropertySources.ProdSource
	}
	if inLocal {
		sources = cloudConfig.PropertySources.LocalSource
	}

	for key, value := range sources {
		viper.Set(key, value)
		fmt.Printf("Loading config property > %s - %s \n", key, value)
	}
	if viper.IsSet(constants.ServiceNameKey) {
		fmt.Println("Successfully loaded configuration for service\n", viper.GetString("serverName"))
	}
}

// Structs having same structure as response from Spring Cloud Config
type springCloudConfig struct {
	Name            string         `json:"name"`
	PropertySources propertySource `json:"propertySources"`
}
type propertySource struct {
	Source      map[string]interface{} `json:"source"`
	ProdSource  map[string]interface{} `json:"prodSource"`
	LocalSource map[string]interface{} `json:"localSource"`
}
