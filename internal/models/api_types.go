package models

type TokenExchangeResponse struct {
	AccessToken string  `json:"access_token"`
	User        Account `json:"user"`
}
