package providers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/basti42/rat-auth-service/internal/models"
	"github.com/basti42/rat-auth-service/internal/system"
	"golang.org/x/oauth2"
)

var githubOAuthConfig = oauth2.Config{
	ClientID:     system.GITHUB_CLIENT_ID,
	ClientSecret: system.GITHUB_CLIENT_SECRET,
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	RedirectURL: system.SERVER_HTTP + "/oauth/callback/github",
	Scopes:      []string{"user:email"},
}

type OAuthGithub struct {
	githubOAuthConfig oauth2.Config
}

func NewOAuthGithub() *OAuthGithub {
	return &OAuthGithub{
		githubOAuthConfig: githubOAuthConfig,
	}
}

func (g *OAuthGithub) GetOAuthConfig() *oauth2.Config {
	return &g.githubOAuthConfig
}

func (g *OAuthGithub) GetUserInfo(accessToken string) (*models.UserInfo, error) {
	url := "https://api.github.com/user"
	userInfo, err := makeHttpCall("GET", url, accessToken)
	if err != nil {
		return nil, fmt.Errorf("http call: %v", err)
	}

	slog.Debug("RAW USER INFO from github:")
	slog.Debug("%v", userInfo)
	slog.Debug("")

	userId, ok := userInfo["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("Invalid user id")
	}
	slog.Debug(fmt.Sprintf("--> user id = %v", userId))

	sub := fmt.Sprintf("%.0f", userId)
	slog.Debug(fmt.Sprintf("--> sum = %v", sub))

	email, ok := userInfo["email"].(string)
	if !ok {
		email = ""
	}
	slog.Debug(fmt.Sprintf("--> email = %v", email))

	avatar, ok := userInfo["avatar_url"].(string)
	if !ok {
		avatar = ""
	}
	slog.Debug(fmt.Sprintf("--> avatar = %v", avatar))

	return &models.UserInfo{
		Email:  email,
		Sub:    sub,
		Avatar: avatar,
	}, nil

}

func makeHttpCall(method, url, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http create request for: '%v'. Error: %v", url, err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http make request for: %v. Error: %v", url, err)
	}
	defer resp.Body.Close()
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("http response. decoding error: %v", err)
	}
	return userInfo, nil
}
