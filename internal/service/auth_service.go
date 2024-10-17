package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/basti42/rat-auth-service/internal/system"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type OAuth interface {
	getOAuthConfig() *oauth2.Config
	getUserInfo(accessToken string) (*UserInfo, error)
}

func (svc *RestService) OauthLogin(c *gin.Context) {
	provider := c.Param("provider")
	var OAuth OAuth
	if provider == "github" {
		OAuth = newOAuthGithub()
		slog.Debug("using GITHUB as oauth id provider")
	} else {
		slog.Error("Invalid provider", "provider", provider)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// other social logins (google, jira, discord, ...) could be integrated here too

	// generate a random state (same implementation as GenerateVerifier)
	state, err := system.GenerateRandomState(32)
	if err != nil {
		slog.Error("Error generating random state", "GenerateRandomState", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// generate a verifier
	verifier := oauth2.GenerateVerifier()

	// store state and verifier
	// expiration date == length of login typing session
	// (how long can the user input their creds, before the sessoin expires)
	// set to 1 minute for now
	dummyToken, err := svc.repo.InsertToken(time.Now().Add(10*time.Minute).Format(time.RFC3339), "", state, verifier)
	if err != nil {
		slog.Error("Error inserting token", "insertToken", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}
	// Redirect user to consent page and ask for permission
	url := OAuth.getOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

	slog.Debug(fmt.Sprintf("stored dummy token for verification: %v", dummyToken))
	slog.Debug(fmt.Sprintf("url for social login: %v", url))

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (svc *RestService) OAuthCalback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	var OAuth OAuth
	if provider == "github" {
		OAuth = newOAuthGithub()
		slog.Debug("using GITHUB as oauth id provider")
	} else {
		slog.Error("Invalid provider", "provider", provider)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// get verifier from state
	// state should be in the database, else there never was a login request
	token, err := svc.repo.GetTokenByState(state)
	if err != nil {
		slog.Error("Error getting token by state", "getTokenByState", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// check if the 1 minute login session is still valid
	// TODO if login session expired think about a reasonable login fow
	expires, err := time.Parse(time.RFC3339, token.Expires)
	if err != nil || time.Now().After(expires) {
		slog.Error("Token expired", "token.Expires", token.Expires)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALLBACK: successfully found found initial token for STATE and VERIFIER")

	// login session is still valid
	// exchange the provided code for an AccessToken with the ID provider
	config := OAuth.getOAuthConfig()
	oauthToken, err := config.Exchange(context.Background(), code, oauth2.VerifierOption(token.Verifier))
	if err != nil {
		slog.Error("Error exchanging code for token", "config.Exchange", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALBACK: exchanged CODE for USER TOKENS")
	slog.Debug(fmt.Sprintf("%v", oauthToken.AccessToken))
	slog.Debug("")

	// use the provided oauthtoken to get user information from ID provider
	userInfo, err := OAuth.getUserInfo(oauthToken.AccessToken)
	if err != nil {
		slog.Error("Error fetching user info", "configProvider.getUserInfo", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALLBACK: gotten USERINFO from ACCESS TOKEN")
	slog.Debug(userInfo.email, userInfo.sub)
	slog.Debug("")

	// email is the unique identifier. Use forom userInfo.Email
	// store the user if it doesn't exist already
	user, err := svc.repo.GetUserByEmail(userInfo.email)
	if err != nil {
		slog.Debug("CALLBACK: SUB as EMAIL not found in USERS DB, creating a new user")
		// user with this email does not exist
		// create a new user
		sub := provider + ":" + userInfo.sub
		user, err = svc.repo.CreateUser(userInfo.email, sub, userInfo.avatar)
		if err != nil {
			slog.Error("Error inserting user", "insertUser", err)
			c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
			return
		}
	}

	slog.Debug("CALLBACK: user created for SUB by unique EMAIL")
	slog.Debug(user.Id, user.Email)
	slog.Debug("")

	// create auth token with a 10 seconds expiration
	token, err = svc.repo.InsertToken(time.Now().Add(10*time.Second).Format(time.RFC3339), user.Id, "", "")
	if err != nil {
		slog.Error("Error inserting token", "insertToken", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALLBACK: created token with 10sec Expiration Date")
	slog.Debug(fmt.Sprintf("%v", token))
	slog.Debug("")

	slog.Debug("CALLBACK: redirecting to CLIENT_URL including token_id")

	// redirect to home page
	// TODO create proper jwt token
	// this token_id is used only temporary,
	// once the client server receives it, it can exchange it for a
	// proper AccessToken using a different auth-service-url, which then
	// returns a properly formatted JWT with payloads
	c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/?token_id="+token.Id)

}
