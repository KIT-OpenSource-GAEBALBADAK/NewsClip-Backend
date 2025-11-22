package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// === [뉴스] 댓글 ===

func CreateNewsComment(comment *models.NewsComment) error {
	return config.DB.Create(comment).Error
}

func GetNewsComments(newsID uint) ([]models.NewsComment, error) {
	var comments []models.NewsComment
	// Preload("User")를 통해 작성자 정보를 JOIN해서 가져옵니다.
	// CreatedAt 내림차순 (최신순) 정렬
	result := config.DB.Preload("User").
		Where("news_id = ?", newsID).
		Order("created_at DESC").
		Find(&comments)
	return comments, result.Error
}

// === [쇼츠] 댓글 ===

func CreateShortComment(comment *models.ShortComment) error {
	return config.DB.Create(comment).Error
}

func GetShortComments(shortID uint) ([]models.ShortComment, error) {
	var comments []models.ShortComment
	result := config.DB.Preload("User").
		Where("short_id = ?", shortID).
		Order("created_at DESC").
		Find(&comments)
	return comments, result.Error
}

// === [커뮤니티] 댓글 ===

func CreatePostComment(comment *models.PostComment) error {
	return config.DB.Create(comment).Error
}

func GetPostComments(postID uint) ([]models.PostComment, error) {
	var comments []models.PostComment
	result := config.DB.Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&comments)
	return comments, result.Error
}
