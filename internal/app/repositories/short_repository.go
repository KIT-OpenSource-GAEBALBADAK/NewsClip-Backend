package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"

	"gorm.io/gorm"
)

// 최신 쇼츠 목록 조회 (커서 페이징 적용)
// cursorID가 0이면 가장 최신부터, 0보다 크면 그 ID보다 작은 것부터 조회
func FindRecentShorts(limit int, cursorID uint) ([]models.Short, error) {
	var shorts []models.Short

	query := config.DB.Model(&models.Short{})

	// [핵심 로직] 커서가 존재하면, 해당 ID보다 작은(오래된) 데이터를 찾음
	if cursorID > 0 {
		query = query.Where("id < ?", cursorID)
	}

	// ID 내림차순 (최신순) 정렬 후 Limit
	result := query.Order("id DESC").Limit(limit).Find(&shorts)

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

// === 쇼츠 상호작용 처리를 위한 5개 함수 ===

// 1. 기존 상호작용 조회
func FindShortInteraction(tx *gorm.DB, userID, shortID uint) (models.ShortInteraction, error) {
	var interaction models.ShortInteraction
	result := tx.Where("user_id = ? AND short_id = ?", userID, shortID).First(&interaction)
	return interaction, result.Error
}

// 2. 상호작용 생성
func CreateShortInteraction(tx *gorm.DB, interaction *models.ShortInteraction) error {
	return tx.Create(interaction).Error
}

// 3. 상호작용 삭제 (취소)
func DeleteShortInteraction(tx *gorm.DB, interaction *models.ShortInteraction) error {
	return tx.Delete(interaction).Error
}

// 4. 상호작용 타입 변경 (like <-> dislike)
func UpdateShortInteraction(tx *gorm.DB, interaction *models.ShortInteraction, newType string) error {
	return tx.Model(interaction).Update("interaction_type", newType).Error
}

// 5. 쇼츠 카운트 업데이트 (LikeCount, DislikeCount 증감)
func UpdateShortCounts(tx *gorm.DB, shortID uint, likeDelta int, dislikeDelta int) error {
	return tx.Model(&models.Short{}).Where("id = ?", shortID).Updates(map[string]interface{}{
		"like_count":    gorm.Expr("like_count + ?", likeDelta),
		"dislike_count": gorm.Expr("dislike_count + ?", dislikeDelta),
	}).Error
}
