package system

import (
	"fmt"
	"log/slog"
	"os"
)

func mustExist(envString string) string {
	env, exists := os.LookupEnv(envString)
	if !exists {
		msg := fmt.Sprintf("environment variable '%v' does not exist.", envString)
		slog.Error(msg)
		panic(msg)
	}
	return env
}

var (
	SERVICE_NAME         = mustExist("SERVICE_NAME")
	SERVER_HTTP          = mustExist("SERVER_HTTP")
	PORT                 = mustExist("PORT")
	DB_HOST              = mustExist("DB_HOST")
	DB_PORT              = mustExist("DB_PORT")
	DB_NAME              = mustExist("DB_NAME")
	DB_USER              = mustExist("DB_USER")
	DB_PASSWORD          = mustExist("DB_PASSWORD")
	LOG_LEVEL            = mustExist("LOG_LEVEL")
	CLIENT_URL           = mustExist("CLIENT_URL")
	GITHUB_CLIENT_ID     = mustExist("GITHUB_CLIENT_ID")
	GITHUB_CLIENT_SECRET = mustExist("GITHUB_CLIENT_SECRET")
	JWT_SECRET           = mustExist("JWT_SECRET")
	GITHUB_CALLBACK_URL  = mustExist("GITHUB_CALLBACK_URL")
)
