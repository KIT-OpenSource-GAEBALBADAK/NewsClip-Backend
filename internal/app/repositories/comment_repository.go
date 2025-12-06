package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"

	"gorm.io/gorm"
)

// === [뉴스] 댓글 ===

// 댓글 생성 시 news 테이블의 comment_count도 +1
func CreateNewsComment(comment *models.NewsComment) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 댓글 생성
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 2. 뉴스 테이블의 comment_count +1 증가
		if err := tx.Model(&models.News{}).
			Where("id = ?", comment.NewsID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func GetNewsComments(newsID uint) ([]models.NewsComment, error) {
	var comments []models.NewsComment
	result := config.DB.Preload("User").
		Where("news_id = ?", newsID).
		Order("created_at DESC").
		Find(&comments)
	return comments, result.Error
}

// === [쇼츠] 댓글 ===

// 댓글 생성 시 shorts 테이블의 comment_count도 +1
func CreateShortComment(comment *models.ShortComment) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 댓글 생성
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 2. 쇼츠 테이블의 comment_count +1 증가
		if err := tx.Model(&models.Short{}).
			Where("id = ?", comment.ShortID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
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

// 댓글 생성 시 posts 테이블의 comment_count도 +1
func CreatePostComment(comment *models.PostComment) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 댓글 생성
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 2. 게시글 테이블의 comment_count +1 증가
		if err := tx.Model(&models.Post{}).
			Where("id = ?", comment.PostID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func GetPostComments(postID uint) ([]models.PostComment, error) {
	var comments []models.PostComment
	result := config.DB.Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&comments)
	return comments, result.Error
}
