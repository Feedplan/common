package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gitlab.com/feedplan-libraries/constants"
)

//Init :
func Init(service, env, path string) {
	addSysConfig()
	body, err := fetchConfiguration(service, path, env)
	if err != nil {
		fmt.Println("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
	}
	parseConfiguration(body)
}

// Make HTTP request to fetch configuration from config server
func fetchConfiguration(service, path, env string) ([]byte, error) {
	var bodyBytes []byte
	var err error
	result := strings.Compare(env, constants.DevEnvironment)
	if result == 0 {
		//panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
		bodyBytes, err = ioutil.ReadFile(path + "/config/config.json")
		if err != nil {
			fmt.Println("Couldn't read local configuration file.", err)
		} else {
			log.Print("using local config.")
		}
	} else {
		url := constants.ConfigURLPath + +service + "/" + env
		fmt.Printf("url is : %s \n", url)
		fmt.Printf("Loading config from %s \n", url)
		resp, err := http.Get(url)
		if resp != nil || err == nil {
			bodyBytes, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading configuration response body.")
			}
		}
	}
	return bodyBytes, err
}

// Get DB and cred from sys env
func addSysConfig() {
	dbUser := getEnvOrDefault("DB_USERNAME", "postgres")
	viper.Set("database.username", dbUser)
	dbPassord := getEnvOrDefault("DB_PASSWORD", "flywaydb")
	viper.Set("database.password", dbPassord)
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	viper.Set("database.host", dbHost)
}

func getEnvOrDefault(envKey, defaultValue string) string {
	var envValue string
	var ok bool
	if envValue, ok = os.LookupEnv(envKey); !ok {
		envValue = defaultValue
	}
	return envValue
}

// Pass JSON bytes into struct and then into Viper
func parseConfiguration(body []byte) {
	var cloudConfig springCloudConfig
	err := json.Unmarshal(body, &cloudConfig)
	if err != nil {
		fmt.Println("Cannot parse configuration, message: " + err.Error())
	}
	for key, value := range cloudConfig.PropertySources[0].Source {
		viper.Set(key, value)
		fmt.Printf("Loading config property > %s - %s \n", key, value)
	}
	if viper.IsSet("server_name") {
		fmt.Println("Successfully loaded configuration for service\n", viper.GetString("server_name"))
	}
}

// Structs having same structure as response from Spring Cloud Config
type springCloudConfig struct {
	Name            string           `json:"name"`
	PropertySources []propertySource `json:"propertySources"`
}
type propertySource struct {
	Name   string                 `json:"name"`
	Source map[string]interface{} `json:"source"`
}
