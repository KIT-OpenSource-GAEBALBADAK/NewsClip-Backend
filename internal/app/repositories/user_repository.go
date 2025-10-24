package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// username으로 유저 조회
func FindUserByUsername(username string) (models.User, error) {
	var user models.User
	result := config.DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

// userID로 유저 조회 (Refresh 시 필요)
func FindUserByID(id uint) (models.User, error) {
	var user models.User
	result := config.DB.First(&user, id)
	return user, result.Error
}

// 유저 생성
func CreateUser(user *models.User) error {
	return config.DB.Create(user).Error
}
