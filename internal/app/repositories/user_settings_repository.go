package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// 기존 설정 조회
func GetUserSettings(userID uint) (models.UserSetting, error) {
	var settings models.UserSetting
	result := config.DB.First(&settings, "user_id = ?", userID)

	// 설정이 없으면 새로 생성 (default 값 사용)
	if result.RowsAffected == 0 {
		settings = models.UserSetting{
			UserID: userID,
		}
		config.DB.Create(&settings)
	}

	return settings, result.Error
}

// 설정 업데이트
func UpdateUserSettings(settings *models.UserSetting, updates map[string]interface{}) error {
	return config.DB.Model(settings).Updates(updates).Error
}
