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

// 게시글 단건 조회 (삭제 권한 확인용)
func FindPostByID(postID uint) (models.Post, error) {
	var post models.Post
	err := config.DB.First(&post, postID).Error
	return post, err
}

// 게시글 삭제
func DeletePost(post *models.Post) error {
	return config.DB.Delete(post).Error
}

// === [7.7] 내가 쓴 게시글 목록 조회 ===
func GetMyPosts(userID uint, page, size int) ([]models.Post, error) {
	var posts []models.Post

	offset := (page - 1) * size

	err := config.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

// === 게시글 ID 목록에 대한 좋아요/싫어요 상태 조회 (Batch Query) ===
func GetPostInteractionsByIDs(userID uint, postIDs []uint) ([]models.PostInteraction, error) {
	var interactions []models.PostInteraction
	// SELECT * FROM post_interactions WHERE user_id = ? AND post_id IN (?, ?, ...)
	err := config.DB.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&interactions).Error
	return interactions, err
}

// (참고: 기존 GetPostList 함수는 Service에서 호출할 때 Preload를 사용하므로 여기엔 없어도 됩니다.
// 만약 Repository 레벨에서 목록을 가져오는 함수가 있다면 Preload("User").Preload("Images")가 포함되어야 합니다.)
func FindPosts(category string, page, size int) ([]models.Post, int64, error) {
	var posts []models.Post
	var totalCount int64

	query := config.DB.Model(&models.Post{})

	if category != "all" && category != "전체" {
		query = query.Where("category = ?", category)
	}

	query.Count(&totalCount)

	offset := (page - 1) * size

	// 작성자(User)와 이미지(Images) 정보를 함께 가져옴 (Eager Loading)
	err := query.Preload("User").Preload("Images").
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&posts).Error

	return posts, totalCount, err
}
