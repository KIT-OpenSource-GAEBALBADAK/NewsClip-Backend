package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"

	"gorm.io/gorm"
)

func GetPostsWithRelations(postType string, page, size int) ([]models.Post, error) {
	var posts []models.Post

	query := config.DB.
		Preload("User").
		Preload("Images").
		Model(&models.Post{})

	if postType != "all" {
		query = query.Where("section = ?", postType)
	}

	offset := (page - 1) * size

	err := query.Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

func CreatePost(post *models.Post) error {
	return config.DB.Create(post).Error
}

func CreatePostImage(postID uint, imageURL string) error {
	return config.DB.Create(&models.PostImage{
		PostID:   postID,
		ImageURL: imageURL,
	}).Error
}

// 1. 기존 상호작용 조회
func FindPostInteraction(tx *gorm.DB, userID, postID uint) (models.PostInteraction, error) {
	var interaction models.PostInteraction
	result := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&interaction)
	return interaction, result.Error
}

// 2. 상호작용 생성
func CreatePostInteraction(tx *gorm.DB, interaction *models.PostInteraction) error {
	return tx.Create(interaction).Error
}

// 3. 상호작용 삭제 (취소)
func DeletePostInteraction(tx *gorm.DB, interaction *models.PostInteraction) error {
	return tx.Delete(interaction).Error
}

// 4. 상호작용 타입 변경
func UpdatePostInteraction(tx *gorm.DB, interaction *models.PostInteraction, newType string) error {
	return tx.Model(interaction).Update("interaction_type", newType).Error
}

// 5. 게시글 카운트 업데이트 (LikeCount, DislikeCount 증감)
func UpdatePostCounts(tx *gorm.DB, postID uint, likeDelta int, dislikeDelta int) error {
	return tx.Model(&models.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"like_count":    gorm.Expr("like_count + ?", likeDelta),
		"dislike_count": gorm.Expr("dislike_count + ?", dislikeDelta),
	}).Error
}
