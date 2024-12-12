package repository

import (
	"github.com/basti42/rat-auth-service/internal/models"
	"github.com/google/uuid"
)

func (repo *AuthRepo) GetUserAccountByEmail(email string) (*models.Account, error) {
	var user models.Account
	tx := repo.db.Model(&models.Account{}).Where("email = ?", email).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func (repo *AuthRepo) GetUserAccountByID(uuid uuid.UUID) (*models.Account, error) {
	var user models.Account
	tx := repo.db.Preload("Subscription").First(&user, uuid)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func (repo *AuthRepo) CreateUserAccount(email, sub, avatar string) (*models.Account, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user := models.Account{
		Id:        id.String(),
		Email:     email,
		Role:      "user",
		Providers: sub,
		Avatar:    avatar,
		Subscription: models.Subscription{
			AccountId:     id.String(),
			Type:          "free",
			NumberOfSeats: 0,
		},
	}
	if tx := repo.db.Create(&user); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}
