package models

import (
	"time"
)

// User: 사용자 기본 정보 (users)
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     *string   `gorm:"type:varchar(30);unique" json:"username"`      // [수정] 포인터 타입
	PasswordHash *string   `gorm:"type:text" json:"-"`                           // [수정] 포인터 타입
	// Name 필드 삭제
	Nickname     *string   `gorm:"type:varchar(30)" json:"nickname"`          // [수정] 포인터 타입
	ProfileImage *string   `gorm:"type:text" json:"profile_image"`          // [수정] 포인터 타입
	Role         string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// === [추가] 소셜 로그인 정보 ===
	Provider   *string `gorm:"type:varchar(50)" json:"provider"`  // 예: "kakao", "google"
	ProviderID *string `gorm:"type:varchar(255);unique" json:"-"` // 소셜 서비스의 유저 고유 ID

	// 관계 설정 (User가 소유한 것들)
	UserSetting   UserSetting    `gorm:"foreignKey:UserID" json:"-"`
	Sessions      []Session      `gorm:"foreignKey:UserID" json:"-"`
	NewsLikes     []NewsLike     `gorm:"foreignKey:UserID" json:"-"`
	NewsBookmarks []NewsBookmark `gorm:"foreignKey:UserID" json:"-"`
	NewsComments  []NewsComment  `gorm:"foreignKey:UserID" json:"-"`
	ShortLikes    []ShortLike    `gorm:"foreignKey:UserID" json:"-"`
	ShortComments []ShortComment `gorm:"foreignKey:UserID" json:"-"`
	Posts         []Post         `gorm:"foreignKey:UserID" json:"-"`
	PostLikes     []PostLike     `gorm:"foreignKey:UserID" json:"-"`
	PostComments  []PostComment  `gorm:"foreignKey:UserID" json:"-"`
	AlertKeywords []AlertKeyword `gorm:"foreignKey:UserID" json:"-"`
	Notifications []Notification `gorm:"foreignKey:UserID" json:"-"`
	Reported      []Report       `gorm:"foreignKey:ReporterID" json:"-"`
	Bans          []Ban          `gorm:"foreignKey:UserID" json:"-"`
}

// UserSetting: 사용자 설정 (user_settings)
type UserSetting struct {
	UserID       uint   `gorm:"primaryKey" json:"-"`
	PushEnabled  bool   `gorm:"default:true" json:"push_enabled"`
	SoundEnabled bool   `gorm:"default:true" json:"sound_enabled"`
	Theme        string `gorm:"type:varchar(20);default:'light'" json:"theme"`
}

// Session: 세션 / 토큰 관리 (sessions)
type Session struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	UserID       uint      `gorm:"not null" json:"-"`
	RefreshToken string    `gorm:"type:text;not null" json:"-"`
	ExpiresAt    time.Time `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"-"`
}
