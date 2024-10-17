package repository

import (
	"github.com/basti42/rat-auth-service/internal/models"
	"github.com/google/uuid"
)

func (repo *AuthRepo) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	tx := repo.db.Model(&models.User{}).Where("email = ?", email).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func (repo *AuthRepo) CreateUser(email, sub, avatar string) (*models.User, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user := models.User{
		Id:     id.String(),
		Email:  email,
		Role:   "user",
		Sub:    sub,
		Avatar: avatar,
	}
	if tx := repo.db.Create(&user); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}
