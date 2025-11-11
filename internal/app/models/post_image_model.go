package models

// PostImage: 커뮤니티 게시글 1:N 이미지
type PostImage struct {
	ID       uint   `gorm:"primaryKey"`
	PostID   uint   `gorm:"not null"`
	ImageURL string `gorm:"type:text;not null"`
}
