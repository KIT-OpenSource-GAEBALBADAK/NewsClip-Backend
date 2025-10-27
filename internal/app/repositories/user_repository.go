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

// === [추가] ===
// Provider와 ProviderID로 유저를 찾습니다.
func FindUserBySocial(provider, providerID string) (models.User, error) {
	var user models.User
	result := config.DB.Where("provider = ? AND provider_id = ?", provider, &providerID).First(&user)
	return user, result.Error
}

// CreateUser 함수를 소셜 로그인에서도 사용할 수 있도록 약간 수정합니다.
func CreateUser(user *models.User) error {
	result := config.DB.Create(user)
	return result.Error
}
