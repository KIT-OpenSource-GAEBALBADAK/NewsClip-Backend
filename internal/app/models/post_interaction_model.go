package models

import "time"

// PostInteraction: 게시글 좋아요/싫어요
type PostInteraction struct {
	UserID          uint   `gorm:"primaryKey"`
	PostID          uint   `gorm:"primaryKey"`
	InteractionType string `gorm:"type:varchar(10);not null"` // 'like' or 'dislike'
	CreatedAt       time.Time
}
