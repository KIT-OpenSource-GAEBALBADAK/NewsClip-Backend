package models

import (
	"time"

	"gorm.io/gorm"
)

// AlertKeyword: 키워드 알림 등록 (alert_keywords)
type AlertKeyword struct {
	gorm.Model
	UserID  uint
	Keyword string `gorm:"type:varchar(50);not null"`
	User    User   // belongs to User
}

// Notification: 알림 이력 (notifications)
type Notification struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Title     string `gorm:"type:varchar(200)"`
	Body      string `gorm:"type:text"`
	Deeplink  string `gorm:"type:text"`
	IsRead    bool   `gorm:"default:false"`
	CreatedAt time.Time
	User      User // belongs to User
}
