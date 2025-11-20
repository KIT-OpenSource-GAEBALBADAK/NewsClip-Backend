package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// 1. 최신 쇼츠 목록 조회 (Limit)
func FindRecentShorts(limit int) ([]models.Short, error) {
	var shorts []models.Short
	// 최신순(created_at DESC)으로 정렬하여 limit만큼 가져옴
	result := config.DB.Order("created_at DESC").Limit(limit).Find(&shorts)
	return shorts, result.Error
}

// 2. (최적화) 특정 유저가 특정 쇼츠 ID 목록들에 대해 어떤 상호작용을 했는지 조회
// (쇼츠 10개를 가져올 때, 상호작용 여부를 확인하기 위해 10번 쿼리하는 것을 방지)
func FindShortInteractionsByIDs(userID uint, shortIDs []uint) ([]models.ShortInteraction, error) {
	var interactions []models.ShortInteraction
	// SELECT * FROM short_interactions WHERE user_id = ? AND short_id IN (?, ?, ...)
	result := config.DB.Where("user_id = ? AND short_id IN ?", userID, shortIDs).Find(&interactions)
	return interactions, result.Error
}

// 3. (참고) 쇼츠 상세 조회 (나중에 필요할 수 있음)
func FindShortByID(shortID uint) (models.Short, error) {
	var short models.Short
	result := config.DB.First(&short, shortID)
	return short, result.Error
}
