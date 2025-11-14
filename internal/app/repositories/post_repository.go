package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
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
