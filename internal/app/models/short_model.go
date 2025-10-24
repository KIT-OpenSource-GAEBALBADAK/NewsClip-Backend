package models

import (
	"time"

	"gorm.io/gorm"
)

// Short: AI 요약 뉴스 (shorts)
type Short struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NewsID    *uint     `json:"news_id"` // 원본 뉴스가 없을 수 있으므로 포인터
	Summary   string    `gorm:"type:text;not null" json:"summary"`
	ImageURL  string    `gorm:"type:text" json:"image_url"`
	LikeCount int       `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time `json:"created_at"`

	// 관계 설정
	News     *News          `json:"-"` // belongs to News
	Likes    []ShortLike    `gorm:"foreignKey:ShortID" json:"-"`
	Comments []ShortComment `gorm:"foreignKey:ShortID" json:"-"`
}

// ShortLike: 쇼츠 좋아요 (short_likes)
type ShortLike struct {
	UserID    uint `gorm:"primaryKey"`
	ShortID   uint `gorm:"primaryKey"`
	CreatedAt time.Time
	User      User  // belongs to User
	Short     Short // belongs to Short
}

// ShortComment: 쇼츠 댓글 (short_comments)
type ShortComment struct {
	gorm.Model
	ShortID uint
	UserID  uint
	Content string `gorm:"type:text;not null"`
	User    User   // belongs to User
	Short   Short  // belongs to Short
}
