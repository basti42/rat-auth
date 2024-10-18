package models

type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}
