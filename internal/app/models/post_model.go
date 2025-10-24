package models

import (
	"time"

	"gorm.io/gorm"
)

// Post: 게시글 (posts)
type Post struct {
	gorm.Model
	UserID       uint
	Title        string `gorm:"type:varchar(200);not null"`
	Content      string `gorm:"type:text;not null"`
	Section      string `gorm:"type:varchar(20);default:'general'"`
	LikeCount    int    `gorm:"default:0"`
	CommentCount int    `gorm:"default:0"`

	// 관계 설정
	User     User          `json:"user"` // 작성자 정보 포함
	Likes    []PostLike    `gorm:"foreignKey:PostID" json:"-"`
	Comments []PostComment `gorm:"foreignKey:PostID" json:"-"`
}

// PostLike: 게시글 좋아요 (post_likes)
type PostLike struct {
	UserID    uint `gorm:"primaryKey"`
	PostID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
	User      User // belongs to User
	Post      Post // belongs to Post
}

// PostComment: 게시글 댓글 (post_comments)
type PostComment struct {
	gorm.Model
	PostID  uint
	UserID  uint
	Content string `gorm:"type:text;not null"`
	User    User   // belongs to User
	Post    Post   // belongs to Post
}
