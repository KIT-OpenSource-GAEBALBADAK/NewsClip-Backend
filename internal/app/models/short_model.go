package models

import (
	"time"
)

// Short: AI 요약 뉴스 (shorts)
type Short struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	NewsID   uint   `json:"news_id"` // 1:1 관계가 명확하므로 포인터 제거 권장 (선택사항)
	Title    string `gorm:"type:varchar(100);not null" json:"title"`
	Summary  string `gorm:"type:text;not null" json:"summary"`
	ImageURL string `gorm:"type:text" json:"image_url"`

	// [수정] 캐시 컬럼: 좋아요, 싫어요 수
	LikeCount    int `gorm:"default:0" json:"like_count"`
	DislikeCount int `gorm:"default:0" json:"dislike_count"` // [신규]

	CreatedAt time.Time `json:"created_at"`

	// 관계 설정
	News News `gorm:"foreignKey:NewsID" json:"news"`

	// [수정] Likes -> Interactions 변경
	Interactions []ShortInteraction `gorm:"foreignKey:ShortID" json:"-"`
	Comments     []ShortComment     `gorm:"foreignKey:ShortID" json:"-"`
}

// [삭제됨] type ShortLike struct { ... }
// -> ShortInteraction이 이 역할을 대신합니다.

// ShortComment: 쇼츠 댓글 (short_comments)
type ShortComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"` // gorm.Model 대신 ID, CreatedAt 등 명시하면 JSON 제어하기 편함
	ShortID   uint      `json:"short_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"` // gorm.Model 포함 시 자동 생성됨

	User User `json:"user"` // 댓글 작성자 정보 포함
}
