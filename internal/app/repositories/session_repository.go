package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// 세션(Refresh Token)을 생성합니다.
func CreateSession(session *models.Session) error {
	result := config.DB.Create(session)
	return result.Error
}

// Refresh Token으로 세션 조회
func FindSessionByToken(token string) (models.Session, error) {
	var session models.Session
	result := config.DB.Where("refresh_token = ?", token).First(&session)
	return session, result.Error
}

// 세션 정보 업데이트 (Token Rotation)
func UpdateSession(session *models.Session) error {
	return config.DB.Save(session).Error
}