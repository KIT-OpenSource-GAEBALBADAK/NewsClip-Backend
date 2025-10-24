package models

import (
	"time"
)

// Report: 통합 신고 테이블 (reports)
type Report struct {
	ID          uint   `gorm:"primaryKey"`
	ContentType string `gorm:"type:varchar(30);not null"`
	ContentID   uint   `gorm:"not null"`
	ReporterID  uint   `gorm:"not null"`
	Reason      string `gorm:"type:text"`
	Status      string `gorm:"type:varchar(20);default:'pending'"`
	CreatedAt   time.Time
	Reporter    User `gorm:"foreignKey:ReporterID"` // belongs to User
}

// Ban: 이용정지 정보 (bans)
type Ban struct {
	ID          uint       `gorm:"primaryKey"`
	UserID      uint       `gorm:"not null"`
	Reason      string     `gorm:"type:text"`
	BannedUntil *time.Time // NULL일 경우 영구 정지
	CreatedAt   time.Time
	User        User // belongs to User
}
