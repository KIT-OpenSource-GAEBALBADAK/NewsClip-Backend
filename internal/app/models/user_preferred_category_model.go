package models

// UserPreferredCategory: P1. 사용자 선호 카테고리
type UserPreferredCategory struct {
	UserID       uint   `gorm:"primaryKey"`
	CategoryName string `gorm:"type:varchar(50);primaryKey"`
}
