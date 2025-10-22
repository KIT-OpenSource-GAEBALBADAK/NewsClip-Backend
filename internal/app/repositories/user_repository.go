package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// username으로 유저를 찾습니다.
func FindUserByUsername(username string) (models.User, error) {
	var user models.User
	result := config.DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

// 유저를 생성합니다.
func CreateUser(user *models.User) error {
	result := config.DB.Create(user)
	return result.Error
}
