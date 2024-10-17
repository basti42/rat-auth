package system

import (
	"fmt"
	"log/slog"
)

func InitLogger() {
	if LOG_LEVEL == "info" {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	slog.Info(fmt.Sprintf("set log level='%v'", LOG_LEVEL))
}
