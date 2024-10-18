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

	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)

	router.GET("/health", app.Health)
	router.GET("/oauth/login/:provider", app.OauthLogin)
	router.GET("/oauth/callback/:provider", app.OauthCallback)
	router.GET("/oauth/exchange/:token_id", app.TokenExchange)

	serviceName := system.SERVICE_NAME
	port := system.PORT
	slog.Info(fmt.Sprintf("starting [%v] on port=%v", serviceName, port))
	if err := router.Run(fmt.Sprintf(":%v", port)); err != nil {
		msg := fmt.Sprintf("error running [%v]: %v", serviceName, err)
		slog.Error(msg)
		panic(msg)
	}
}
