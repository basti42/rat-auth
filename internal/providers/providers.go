package providers

import (
	"github.com/basti42/rat-auth-service/internal/models"
	"golang.org/x/oauth2"
)

// OAuth is the interface used for individual providers
// each individual provider is implemented in its own file within the providers package
type OAuth interface {
	GetOAuthConfig() *oauth2.Config
	GetUserInfo(accessToken string) (*models.UserInfo, error)
}
