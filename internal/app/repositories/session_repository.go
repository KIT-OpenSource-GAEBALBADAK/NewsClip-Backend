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

// (참고) 나중에 리프레시 토큰으로 세션을 찾거나 삭제하는 함수도 여기에 추가됩니다.
