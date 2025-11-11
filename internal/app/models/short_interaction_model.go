package models

import "time"

// ShortInteraction: 쇼츠 좋아요/싫어요 (P3)
type ShortInteraction struct {
	UserID          uint   `gorm:"primaryKey"`
	ShortID         uint   `gorm:"primaryKey"`
	InteractionType string `gorm:"type:varchar(10);not null"` // 'like' or 'dislike'
	CreatedAt       time.Time
}
