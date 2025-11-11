package models

import (
	"time"

	"gorm.io/gorm"
)

// News: 뉴스 본문 데이터 (news)
// News: 뉴스 본문 (캐시 컬럼 추가)
type News struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ExternalID  string    `gorm:"type:varchar(255);unique" json:"external_id"`
	Title       string    `gorm:"type:text;not null" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	Source      string    `gorm:"type:text" json:"source"`
	URL         string    `gorm:"type:text" json:"url"`
	Category    string    `gorm:"type:varchar(50)" json:"category"`
	ImageURL    string    `gorm:"type:text" json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`   // DB 저장 시간
	PublishedAt time.Time `json:"published_at"` // 원본 기사 시간

	// === [신규] 캐시/UI용 컬럼 ===
	ViewCount       int `gorm:"default:0" json:"view_count"`
	LikeCount       int `gorm:"default:0" json:"like_count"`
	DislikeCount    int `gorm:"default:0" json:"dislike_count"`
	CommentCount    int `gorm:"default:0" json:"comment_count"`
	ReadTimeMinutes int `gorm:"default:3" json:"read_time_minutes"`

	// 관계 설정 (Interaction, Comment)
	Interactions []NewsInteraction `gorm:"foreignKey:NewsID" json:"-"`
	Comments     []NewsComment     `gorm:"foreignKey:NewsID" json:"-"`
	Bookmarks    []NewsBookmark    `gorm:"foreignKey:NewsID" json:"-"`
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
