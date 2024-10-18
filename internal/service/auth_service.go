package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/basti42/rat-auth-service/internal/models"
	"github.com/basti42/rat-auth-service/internal/providers"
	"github.com/basti42/rat-auth-service/internal/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func (svc *RestService) OauthLogin(c *gin.Context) {
	provider := c.Param("provider")
	var OAuth providers.OAuth
	if provider == "github" {
		OAuth = providers.NewOAuthGithub()
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
	loginSessionToken, err := svc.repo.InsertToken(time.Now().Add(10*time.Minute).Format(time.RFC3339), "", state, verifier, "", models.LoginSessionToken)
	if err != nil {
		slog.Error("Error inserting token", "insertToken", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}
	// Redirect user to consent page and ask for permission
	url := OAuth.GetOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

	slog.Debug(fmt.Sprintf("stored dummy token for verification: %v", loginSessionToken))
	slog.Debug(fmt.Sprintf("url for social login: %v", url))

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (svc *RestService) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	var OAuth providers.OAuth
	if provider == "github" {
		OAuth = providers.NewOAuthGithub()
		slog.Debug("using GITHUB as oauth id provider")
	} else {
		slog.Error("Invalid provider", "provider", provider)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// get verifier from state
	// state should be in the database, else there never was a login request
	exchangeSessionToken, err := svc.repo.GetTokenByState(state)
	if err != nil {
		slog.Error("Error getting token by state", "getTokenByState", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// check if the 1 minute login session is still valid
	// TODO if login session expired think about a reasonable login fow
	expires, err := time.Parse(time.RFC3339, exchangeSessionToken.Expires)
	if err != nil || time.Now().After(expires) {
		slog.Error("Token expired", "token.Expires", exchangeSessionToken.Expires)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALLBACK: successfully found found initial token for STATE and VERIFIER")

	// login session is still valid
	// exchange the provided code for an AccessToken with the ID provider
	// this is just an internal Access Token to be used for accessing the
	// ID providers resources, for synching the users infos, DO NOT USE FOR APP AUTH
	config := OAuth.GetOAuthConfig()
	oauthToken, err := config.Exchange(context.Background(), code, oauth2.VerifierOption(exchangeSessionToken.Verifier))
	if err != nil {
		slog.Error("Error exchanging code for token", "config.Exchange", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	// use the provided oauthtoken to get user information from ID provider
	userInfo, err := OAuth.GetUserInfo(oauthToken.AccessToken)
	if err != nil {
		slog.Error("Error fetching user info", "configProvider.getUserInfo", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug(oauthToken.AccessToken)
	slog.Debug("CALLBACK: gotten USERINFO from ACCESS TOKEN")
	slog.Debug(userInfo.Email, userInfo.Sub)
	slog.Debug("")

	// email is the unique identifier. Use forom userInfo.Email
	// store the user if it doesn't exist already
	user, err := svc.repo.GetUserByEmail(userInfo.Email)
	if err != nil {
		slog.Debug("CALLBACK: SUB as EMAIL not found in USERS DB, creating a new user")
		// user with this email does not exist
		// create a new user
		sub := provider + ":" + userInfo.Sub
		user, err = svc.repo.CreateUser(userInfo.Email, sub, userInfo.Avatar)
		if err != nil {
			slog.Error("Error inserting user", "insertUser", err)
			c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
			return
		}
	}

	slog.Debug("CALLBACK: user created for SUB by unique EMAIL")
	slog.Debug(user.Id, user.Email)
	slog.Debug("")

	exchangeCode := oauth2.GenerateVerifier()

	// create auth token with a 10 seconds expiration
	exchangeSessionToken, err = svc.repo.InsertToken(time.Now().Add(10*time.Minute).Format(time.RFC3339), user.Id, "", "", exchangeCode, models.ExchangeSessionToken)
	if err != nil {
		slog.Error("Error inserting token", "insertToken", err)
		c.Redirect(http.StatusTemporaryRedirect, system.CLIENT_URL+"/auth?error=unauthorized")
		return
	}

	slog.Debug("CALLBACK: created token with 10sec Expiration Date")
	slog.Debug(fmt.Sprintf("%v", exchangeSessionToken))
	slog.Debug("")

	slog.Debug("CALLBACK: redirecting to CLIENT_URL including token_id")

	// redirect to home page
	// TODO create proper jwt token
	// this token_id is used only temporary,
	// once the client server receives it, it can exchange it for a
	// proper AccessToken using a different auth-service-url, which then
	// returns a properly formatted JWT with payloads
	c.Redirect(http.StatusTemporaryRedirect,
		system.CLIENT_URL+"/?token_id="+exchangeSessionToken.Id+"&exc="+exchangeCode)
}

func (svc *RestService) TokenExchange(c *gin.Context) (models.TokenExchangeResponse, error) {
	tokenID := c.Param("token_id")
	exchangeCode := c.Query("exc")

	tokenUUID, err := uuid.Parse(tokenID)
	if err != nil {
		msg := fmt.Sprintf("invalid token id='%v'", tokenID)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	if exchangeCode == "" {
		msg := fmt.Sprintf("mising exchange code, unable to get access token for id='%v'", tokenID)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	token, err := svc.repo.GetTokenByIdExchangeCode(tokenUUID, exchangeCode)
	if err != nil {
		msg := fmt.Sprintf("error getting token for id='%v' and exchange_code='%v'", tokenUUID, exchangeCode)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	exchangeTime, _ := time.Parse(time.RFC3339, token.Expires)
	if time.Now().After(exchangeTime) {
		msg := fmt.Sprintf("token exchange session expired for token id='%v'", tokenID)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	userUUID, err := uuid.Parse(token.UserId)
	if err != nil {
		msg := fmt.Sprintf("invalid user uuid with %v='%v'", token.Type, token.Id)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	user, err := svc.repo.GetUserByID(userUUID)
	if err != nil {
		msg := fmt.Sprintf("no user found for exchange token='%v'", token.Id)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	// access token has a 30 day expiration
	accessToken, err := svc.repo.InsertToken(time.Now().Add(43200*time.Minute).Format(time.RFC3339), token.UserId, "", "", "", models.AccessToken)
	if err != nil {
		msg := fmt.Sprintf("error creating access token for exchange token='%v'", tokenID)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	jwt, err := accessToken.GenerateJWT(user)
	if err != nil {
		msg := fmt.Sprintf("error generating JWT for access token='%v'", accessToken.Id)
		slog.Error(msg)
		return models.TokenExchangeResponse{}, errors.New(msg)
	}

	return models.TokenExchangeResponse{
		AccessToken: jwt,
		User:        *user,
	}, nil
}
