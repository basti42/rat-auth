package service

import (
	"github.com/basti42/rat-auth-service/internal/repository"
	"gorm.io/gorm"
)

type RestService struct {
	repo *repository.AuthRepo
}

func NewRestService(db *gorm.DB) *RestService {
	return &RestService{repo: repository.NewAuthRepo(db)}
}
