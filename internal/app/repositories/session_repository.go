package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// 세션 생성
func CreateSession(session *models.Session) error {
	return config.DB.Create(session).Error
}

// RefreshToken 으로 세션 조회
func FindSessionByToken(token string) (models.Session, error) {
	var session models.Session
	result := config.DB.Where("refresh_token = ?", token).First(&session)
	return session, result.Error
}

// 세션 업데이트 (RefreshToken 재발급 시 갱신)
func UpdateSession(session *models.Session) error {
	return config.DB.Save(session).Error
}
