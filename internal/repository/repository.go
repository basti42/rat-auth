package repository

import (
	"fmt"

	"log/slog"

	"github.com/basti42/rat-auth-service/internal/models"
	"github.com/basti42/rat-auth-service/internal/system"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDBConnection() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		system.DB_HOST, system.DB_USER, system.DB_PASSWORD, system.DB_NAME, system.DB_PORT)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		msg := fmt.Sprintf("error creating auth db: %v", err)
		slog.Error(msg)
		panic(msg)
	}

	if err = db.AutoMigrate(&models.Token{}, &models.User{}); err != nil {
		msg := fmt.Sprintf("error migrating auth db: %v", err)
		slog.Error(msg)
		panic(msg)
	}
	return db
}

type AuthRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

func (repo *AuthRepo) InsertToken(expires, userid, state, verifier, exchangeCode string, tokenType models.TokenType) (*models.Token, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	token := models.Token{
		Id:           id.String(),
		Expires:      expires,
		UserId:       userid,
		State:        state,
		Verifier:     verifier,
		ExchangeCode: exchangeCode,
		Type:         tokenType,
	}
	tx := repo.db.Create(&token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &token, nil
}

func (repo *AuthRepo) GetTokenByState(state string) (*models.Token, error) {
	var token models.Token
	tx := repo.db.Model(&models.Token{}).Where("state = ?", state).First(&token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &token, nil
}

func (repo *AuthRepo) GetTokenByIdExchangeCode(uuid uuid.UUID, exchangeCode string) (*models.Token, error) {
	var token models.Token
	tx := repo.db.Model(&models.Token{}).
		Where("id = ? AND exchange_code = ?", uuid, exchangeCode).
		First(&token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &token, nil
}
