package main

import (
	"fmt"
	"log/slog"

	"github.com/basti42/rat-auth-service/internal"
	"github.com/basti42/rat-auth-service/internal/repository"
	"github.com/basti42/rat-auth-service/internal/system"
	"github.com/gin-gonic/gin"
)

func main() {

	system.InitLogger()

	db := repository.GetDBConnection()

	app := internal.NewApplication(db)

	router := gin.Default()

	router.GET("/health", app.Health)
	router.GET("/oauth/login/:provider", app.OauthLogin)
	router.GET("/oauth/callback/:provider", app.OauthCallback)

	serviceName := system.SERVICE_NAME
	port := system.PORT
	slog.Info(fmt.Sprintf("starting [%v] on port=%v", serviceName, port))
	if err := router.Run(fmt.Sprintf(":%v", port)); err != nil {
		msg := fmt.Sprintf("error running [%v]: %v", serviceName, err)
		slog.Error(msg)
		panic(msg)
	}
}
