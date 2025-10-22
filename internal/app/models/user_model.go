package models

import (
	"time"
)

// User: 사용자 기본 정보 (users)
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(30);unique;not null" json:"username"`
	PasswordHash string    `gorm:"type:text;not null" json:"-"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"`
	Nickname     string    `gorm:"type:varchar(30)" json:"nickname"`
	ProfileImage string    `gorm:"type:text" json:"profile_image"`
	Role         string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

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
