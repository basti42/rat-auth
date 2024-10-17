package internal

import (
	"net/http"

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
	// provider := c.Param(("provider"))
	// slog.Info(fmt.Sprintf("Received request for login via '%v'", provider))

	// gothic.BeginAuthHandler(c.Writer, c.Request)

	service.NewRestService(a.db).OauthLogin(c)
}

func (a *Application) OauthCallback(c *gin.Context) {

	// user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	// if err != nil {
	// 	slog.Error("error completing user authentication")
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("error completing user authentication"))
	// 	return
	// }
	// slog.Info(user.UserID)
	// slog.Info(user.AccessToken)
	// slog.Info("")

	// c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL)
	service.NewRestService(a.db).OAuthCalback(c)
}
