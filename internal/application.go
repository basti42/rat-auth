package internal

import (
	"fmt"
	"net/http"

	"log/slog"

	"github.com/basti42/rat-auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	db *gorm.DB
}

func NewApplication(db *gorm.DB) *Application {
	return &Application{db: db}
}

func (a *Application) Health(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func (a *Application) OauthLogin(c *gin.Context) {
	service.NewRestService(a.db).OauthLogin(c)
}

func (a *Application) OauthCallback(c *gin.Context) {
	service.NewRestService(a.db).OAuthCallback(c)
}

func (a *Application) TokenExchange(c *gin.Context) {
	accessTokenResponse, err := service.NewRestService(a.db).TokenExchange(c)
	if err != nil {
		slog.Warn(fmt.Sprintf("error exchanging token_id with exchange_code for access token"))
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, accessTokenResponse)
}
