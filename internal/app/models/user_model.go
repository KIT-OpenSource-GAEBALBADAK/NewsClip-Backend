package models

import (
	"time"
)

// User: 사용자 기본 정보 (users)
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     *string   `gorm:"type:varchar(30);unique" json:"username"`
	PasswordHash *string   `gorm:"type:text" json:"-"`
	Nickname     *string   `gorm:"type:varchar(30)" json:"nickname"`
	ProfileImage *string   `gorm:"type:text" json:"profile_image"`
	Role         string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 소셜 로그인 정보
	Provider   *string `gorm:"type:varchar(50)" json:"provider"`
	ProviderID *string `gorm:"type:varchar(255);unique" json:"-"`

	// === [관계 설정 수정] ===

	// 1. 설정 및 세션 (기존 유지)
	UserSetting UserSetting `gorm:"foreignKey:UserID" json:"-"`
	Sessions    []Session   `gorm:"foreignKey:UserID" json:"-"`

	// 2. [신규] 선호 카테고리 (P1 추천용)
	PreferredCategories []UserPreferredCategory `gorm:"foreignKey:UserID" json:"-"`

	// 3. [수정] 뉴스 상호작용 (Like -> Interaction)
	NewsInteractions []NewsInteraction `gorm:"foreignKey:UserID" json:"-"`
	NewsBookmarks    []NewsBookmark    `gorm:"foreignKey:UserID" json:"-"`
	NewsComments     []NewsComment     `gorm:"foreignKey:UserID" json:"-"`

	// 4. [수정] 쇼츠 상호작용 (Like -> Interaction)
	ShortInteractions []ShortInteraction `gorm:"foreignKey:UserID" json:"-"`
	ShortComments     []ShortComment     `gorm:"foreignKey:UserID" json:"-"`

	// 5. 커뮤니티 및 기타 (기존 유지)
	Posts []Post `gorm:"foreignKey:UserID" json:"-"`
	// PostLikes     []PostLike     `gorm:"foreignKey:UserID" json:"-"`
	PostComments  []PostComment  `gorm:"foreignKey:UserID" json:"-"`
	AlertKeywords []AlertKeyword `gorm:"foreignKey:UserID" json:"-"`
	Notifications []Notification `gorm:"foreignKey:UserID" json:"-"`
	Reported      []Report       `gorm:"foreignKey:ReporterID" json:"-"`
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
