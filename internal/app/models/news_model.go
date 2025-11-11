package models

import (
	"time"

	"gorm.io/gorm"
)

// News: 뉴스 본문 데이터 (news)
type News struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ExternalID string    `gorm:"type:varchar(255);unique" json:"external_id"`
	Title      string    `gorm:"type:text;not null" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	Source     string    `gorm:"type:text" json:"source"` // 여기에 '연합뉴스' 등 언론사명 저장
	URL        string    `gorm:"type:text" json:"url"`
	Category   string    `gorm:"type:varchar(50)" json:"category"`
	ImageURL   string    `gorm:"type:text" json:"image_url"`
	CreatedAt  time.Time `json:"created_at"` // DB에 저장된 시간 (GORM)
	// === [신규] 원본 기사 작성 시간 ===
	PublishedAt time.Time `json:"published_at"` // Naver API의 pubDate

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
