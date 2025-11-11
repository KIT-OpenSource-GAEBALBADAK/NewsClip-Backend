package models

import "time"

// NewsInteraction: 뉴스 좋아요/싫어요 (P3)
type NewsInteraction struct {
	UserID          uint   `gorm:"primaryKey"`
	NewsID          uint   `gorm:"primaryKey"`
	InteractionType string `gorm:"type:varchar(10);not null"` // 'like' or 'dislike'
	CreatedAt       time.Time
}
