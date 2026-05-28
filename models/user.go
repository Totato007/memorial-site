package models

import (
	"time"
)

// User 用户模型 — 普通用户与管理员共用，通过 Role 字段区分
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Phone        string     `gorm:"uniqueIndex;size:20;not null" json:"phone"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Nickname     string     `gorm:"size:64" json:"nickname"`
	AvatarPath   string     `gorm:"size:255" json:"avatar_path"`
	Bio          string     `gorm:"type:text" json:"bio"`
	Role         string     `gorm:"size:16;default:user" json:"role"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastActiveAt *time.Time `json:"last_active_at"` // 最后活跃时间 (在线状态判定)
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
