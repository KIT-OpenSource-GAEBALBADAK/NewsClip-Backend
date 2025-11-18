package services

import (
	"errors"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

type UpdateNotificationSettingsRequest struct {
	PushEnabled  *bool    `json:"pushEnabled"`
	SoundEnabled *bool    `json:"soundEnabled"`
	Keywords     []string `json:"keywords"`
}

func UpdateNotificationSettings(userID uint, req UpdateNotificationSettingsRequest) error {

	var setting models.UserSetting
	if err := config.DB.Where("user_id = ?", userID).First(&setting).Error; err != nil {
		return errors.New("설정 정보를 찾을 수 없습니다.")
	}

	updates := map[string]interface{}{}

	if req.PushEnabled != nil {
		updates["push_enabled"] = *req.PushEnabled
	}
	if req.SoundEnabled != nil {
		updates["sound_enabled"] = *req.SoundEnabled
	}

	// === Keyword 배열 저장 방식 ===
	// alert_keywords 테이블 싹 삭제 → 다시 입력
	if req.Keywords != nil {
		// 삭제
		config.DB.Where("user_id = ?", userID).Delete(&models.AlertKeyword{})

		// 재삽입
		for _, kw := range req.Keywords {
			config.DB.Create(&models.AlertKeyword{
				UserID:  userID,
				Keyword: kw,
			})
		}
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&setting).Updates(updates).Error; err != nil {
			return errors.New("알림 설정 저장 실패")
		}
	}

	return nil
}
