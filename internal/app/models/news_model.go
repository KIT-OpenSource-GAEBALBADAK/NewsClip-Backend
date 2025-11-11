package models

import (
	"time"

	"gorm.io/gorm"
)

// News: 뉴스 본문 데이터 (news)
type News struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ExternalID string `gorm:"type:varchar(255);unique" json:"external_id"`
	Title      string `gorm:"type:text;not null" json:"title"`
	Content    string `gorm:"type:text" json:"content"`
	// === [수정] ===
	// Varchar(100)은 원문 링크를 저장하기에 너무 작습니다.
	Source   string `gorm:"type:text" json:"source"` // "varchar(100)" -> "text"로 변경
	URL      string `gorm:"type:text" json:"url"`
	Category string `gorm:"type:varchar(50)" json:"category"`
	// === [신규] 이미지 URL 컬럼 추가 ===
	ImageURL  string    `gorm:"type:text" json:"image_url"`
	CreatedAt time.Time `json:"created_at"`

	// 관계 설정
	Likes     []NewsLike     `gorm:"foreignKey:NewsID" json:"-"`
	Bookmarks []NewsBookmark `gorm:"foreignKey:NewsID" json:"-"`
	Comments  []NewsComment  `gorm:"foreignKey:NewsID" json:"-"`
	Shorts    []Short        `gorm:"foreignKey:NewsID" json:"-"`
}

// NewsLike: 뉴스 좋아요 관계 (news_likes)
type NewsLike struct {
	UserID    uint `gorm:"primaryKey"`
	NewsID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
	User      User // belongs to User
	News      News // belongs to News
}

// NewsBookmark: 뉴스 북마크 (news_bookmarks)
type NewsBookmark struct {
	UserID    uint `gorm:"primaryKey"`
	NewsID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
	User      User // belongs to User
	News      News // belongs to News
}

// NewsComment: 뉴스 댓글 (news_comments)
type NewsComment struct {
	gorm.Model
	NewsID  uint
	UserID  uint
	Content string `gorm:"type:text;not null"`
	User    User   // belongs to User
	News    News   // belongs to News
}
