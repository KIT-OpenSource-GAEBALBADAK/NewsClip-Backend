package repositories

import (
	"math"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"time"

	"gorm.io/gorm"
)

// ExternalID(Naver News 링크)로 뉴스를 찾습니다.
// (뉴스 수집 시 중복 체크용)
func FindNewsByExternalID(externalID string) (models.News, error) {
	var news models.News
	result := config.DB.Where("external_id = ?", externalID).First(&news)
	return news, result.Error
}

// 수집된 뉴스 목록을 DB에 일괄 생성(Batch Create)합니다.
func CreateNewsBatch(newsList []models.News) error {
	// GORM의 CreateInBatches를 사용하면 효율적입니다.
	// (단, GORM 2.0 이상 필요)
	result := config.DB.CreateInBatches(newsList, 100) // 100개씩 나눠서 삽입
	return result.Error
}

// === 카테고리별 뉴스 목록 조회 (페이징 포함) ===
// (totalPages 반환을 위해 int64(totalCount)도 함께 반환)
func GetNewsByCategory(category string, page int, size int) ([]models.News, int64, int, error) {
	var newsList []models.News
	var totalCount int64

	// 1. (DB 트랜잭션)
	//    전체 카운트와 목록 조회를 트랜잭션으로 묶어 데이터 일관성 보장
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1-1. category가 "전체"일 경우
		query := tx.Model(&models.News{})
		if category != "전체" {
			query = query.Where("category = ?", category)
		}

		// 1-2. 전체 아이템 개수(totalCount) 조회
		if err := query.Count(&totalCount).Error; err != nil {
			return err
		}

		// 1-3. 페이징 계산 (Offset)
		offset := (page - 1) * size

		// 1-4. 실제 데이터 목록 조회 (최신순: published_at 기준)
		if err := query.Order("published_at DESC").
			Limit(size).
			Offset(offset).
			Find(&newsList).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, 0, 0, err
	}

	// 2. 전체 페이지 수(totalPages) 계산
	totalPages := int(math.Ceil(float64(totalCount) / float64(size)))

	return newsList, totalCount, totalPages, nil
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

// === Primary Key(ID)로 뉴스 1건 조회 ===
func FindNewsByID(newsID uint) (models.News, error) {
	var news models.News
	// ID로 조회
	result := config.DB.First(&news, newsID)
	return news, result.Error
}

// === 뉴스의 조회수(view_count)를 1 증가시킴 ===
func IncrementNewsViewCount(newsID uint) error {
	// GORM의 UpdateColumn을 사용하여 특정 컬럼만 +1 업데이트
	// (SQL: UPDATE news SET view_count = view_count + 1 WHERE id = ?)
	result := config.DB.Model(&models.News{}).Where("id = ?", newsID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	return result.Error
}

// === 상호작용 처리를 위한 5개 함수 ===

// FindNewsInteraction (트랜잭션용)
// : 유저가 해당 뉴스에 대해 기존에 한 상호작용을 찾습니다.
func FindNewsInteraction(tx *gorm.DB, userID, newsID uint) (models.NewsInteraction, error) {
	var interaction models.NewsInteraction
	result := tx.Where("user_id = ? AND news_id = ?", userID, newsID).First(&interaction)
	return interaction, result.Error
}

// CreateNewsInteraction (트랜잭션용)
// : 새로운 상호작용 레코드를 생성합니다.
func CreateNewsInteraction(tx *gorm.DB, interaction *models.NewsInteraction) error {
	return tx.Create(interaction).Error
}

// DeleteNewsInteraction (트랜잭션용)
// : 기존 상호작용 레코드를 삭제합니다. (취소)
func DeleteNewsInteraction(tx *gorm.DB, interaction *models.NewsInteraction) error {
	return tx.Delete(interaction).Error
}

// UpdateNewsInteraction (트랜잭션용)
// : 기존 상호작용 타입을 변경합니다. (like -> dislike)
func UpdateNewsInteraction(tx *gorm.DB, interaction *models.NewsInteraction, newType string) error {
	return tx.Model(interaction).Update("interaction_type", newType).Error
}

// UpdateNewsCounts (트랜잭션용)
// : 'news' 테이블의 캐시 카운트를 증감시킵니다. (Deltas: +1, -1, 0)
func UpdateNewsCounts(tx *gorm.DB, newsID uint, likeDelta int, dislikeDelta int) error {
	return tx.Model(&models.News{}).Where("id = ?", newsID).Updates(map[string]interface{}{
		"like_count":    gorm.Expr("like_count + ?", likeDelta),
		"dislike_count": gorm.Expr("dislike_count + ?", dislikeDelta),
	}).Error
}

// === 북마크 처리를 위한 3개 함수 ===

// FindBookmark (UserID, NewsID로 북마크 조회)
func FindBookmark(userID, newsID uint) (models.NewsBookmark, error) {
	var bookmark models.NewsBookmark
	result := config.DB.Where("user_id = ? AND news_id = ?", userID, newsID).First(&bookmark)
	return bookmark, result.Error
}

// CreateBookmark (새 북마크 생성)
func CreateBookmark(bookmark *models.NewsBookmark) error {
	return config.DB.Create(bookmark).Error
}

// DeleteBookmark (기존 북마크 삭제)
func DeleteBookmark(bookmark *models.NewsBookmark) error {
	return config.DB.Delete(bookmark).Error
}

// === [신규] 사용자가 북마크한 뉴스 목록 조회 (JOIN 및 페이징) ===
func GetBookmarkedNews(userID uint, page int, size int) ([]models.News, int64, int, error) {
	var newsList []models.News
	var totalCount int64

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 전체 북마크 개수
		if err := tx.Model(&models.NewsBookmark{}).
			Where("user_id = ?", userID).
			Count(&totalCount).Error; err != nil {
			return err
		}

		if totalCount == 0 {
			return nil
		}

		offset := (page - 1) * size

		// JOIN 조회
		if err := tx.Table("news AS n").
			Joins("JOIN news_bookmarks AS b ON n.id = b.news_id").
			Where("b.user_id = ?", userID).
			Order("b.created_at DESC").
			Limit(size).
			Offset(offset).
			Find(&newsList).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(size)))

	return newsList, totalCount, totalPages, nil
}
