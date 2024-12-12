package models

import (
	"errors"
	"time"

	"github.com/basti42/rat-auth-service/internal/system"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type TokenType string

var (
	LoginSessionToken    TokenType = "login_session"
	ExchangeSessionToken TokenType = "exchange_session"
	AccessToken          TokenType = "access_token"
)

type Token struct {
	Id           string    `gorm:"primaryKey" json:"id"`
	Expires      string    `gorm:"not null" json:"expires"`
	UserId       string    `gorm:"null" json:"user_id"`
	State        string    `gorm:"null" json:"state"`
	Verifier     string    `gorm:"null" json:"verifier"`
	ExchangeCode string    `gorm:"null" json:"exchange_code"`
	Type         TokenType `gorm:"not null" json:"type"`
}

func (t *Token) GenerateJWT(payload *Account) (string, error) {

	expirationDate, _ := time.Parse(time.RFC3339, t.Expires)

	customClaims := jwt.MapClaims{
		"data": map[string]string{
			"user_uuid":         payload.Id,
			"role":              payload.Role,
			"email":             payload.Email,
			"avatar":            payload.Avatar,
			"subscription_type": payload.Subscription.Type,
		},
		"exp": jwt.NewNumericDate(expirationDate),
		"iat": jwt.NewNumericDate(time.Now()),
		"iss": "remote-agile-toolbox",
		"sub": payload.Email,
		"id":  payload.Id,
		"aud": []string{"rat-services"},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	signedTokenString, err := jwtToken.SignedString([]byte(system.JWT_SECRET))
	if err != nil {
		return "", errors.New("errors signing JWT access key")
	}
	return signedTokenString, nil
}

type CustomClaims struct {
	UserUUID string `json:"user_uuid"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	jwt.RegisteredClaims
}

type Account struct {
	Id           string       `gorm:"primaryKey" json:"id"`
	Email        string       `json:"email"`
	Role         string       `json:"role"`
	Providers    string       `json:"providers"`
	Avatar       string       `json:"avatar"`
	Subscription Subscription `gorm:"foreignKey:AccountId"  json:"subscription"`
}

type Subscription struct {
	gorm.Model
	AccountId     string
	Type          string `json:"type"`
	NumberOfSeats int    `json:"number_of_seats"`
}

type UserInfo struct {
	Email  string
	Sub    string
	Avatar string
}
