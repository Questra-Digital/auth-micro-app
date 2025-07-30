package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// App
	AppEnv              string
	ResourceServicePort int
}

var AppConfig Config

func InitConfig() {
	var err error

	AppConfig = Config{
		AppEnv: getEnv("APP_ENV", "development"),
	}

	// Parse Resource Service Port
	AppConfig.ResourceServicePort, err = parseEnvInt("RESOURCE_SERVICE_PORT", 8084)
	if err != nil {
		log.Fatalf("Invalid RESOURCE_SERVICE_PORT: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseEnvInt(key string, defaultVal int) (int, error) {
	if val := os.Getenv(key); val != "" {
		return strconv.Atoi(val)
	}
	return defaultVal, nil
}
 