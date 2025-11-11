package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"time"
)

// [신규] ExternalID(Naver News 링크)로 뉴스를 찾습니다.
// (뉴스 수집 시 중복 체크용)
func FindNewsByExternalID(externalID string) (models.News, error) {
	var news models.News
	result := config.DB.Where("external_id = ?", externalID).First(&news)
	return news, result.Error
}

// [신규] 수집된 뉴스 목록을 DB에 일괄 생성(Batch Create)합니다.
func CreateNewsBatch(newsList []models.News) error {
	// GORM의 CreateInBatches를 사용하면 효율적입니다.
	// (단, GORM 2.0 이상 필요)
	result := config.DB.CreateInBatches(newsList, 100) // 100개씩 나눠서 삽입
	return result.Error
}

// [참고] 카테고리별 뉴스 목록 조회 (API 명세서 3.1)
// (나중에 컨트롤러에서 사용할 함수 예시)
func GetNewsByCategory(category string, page int, size int) ([]models.News, int64, error) {
	var newsList []models.News
	var total int64

	// 1. 카테고리에 해당하는 전체 뉴스 개수 카운트
	config.DB.Model(&models.News{}).Where("category = ?", category).Count(&total)

	// 2. 페이징 계산 (Offset)
	offset := (page - 1) * size

	// 3. 데이터 조회 (최신순 정렬)
	result := config.DB.Where("category = ?", category).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&newsList)

	return newsList, total, result.Error
}

// === 특정 날짜 이전의 뉴스를 삭제합니다. ===
func DeleteNewsOlderThan(cutoffDate time.Time) (int64, error) {
	// GORM을 사용하여 created_at이 cutoffDate보다 오래된 레코드를 삭제
	result := config.DB.Where("created_at < ?", cutoffDate).Delete(&models.News{})

	if result.Error != nil {
		return 0, result.Error
	}

	// 삭제된 행(row)의 개수를 반환
	return result.RowsAffected, nil
}
