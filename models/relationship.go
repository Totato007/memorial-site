package models

import "time"

// Relationship 关系模型 — 用户对好友的单向绑定
// 约束: 同一 (UserID, FriendID, Type) 组合唯一
// 业务规则: Type=couple 时，每个用户最多一条活跃记录
type Relationship struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index:idx_rel_user,unique" json:"user_id"`
	FriendID   uint       `gorm:"not null;index:idx_rel_friend,unique" json:"friend_id"`
	Type       string     `gorm:"size:32;default:friend" json:"type"` // couple|friend|family|custom
	CustomType string     `gorm:"size:64" json:"custom_type"`
	Nickname   string     `gorm:"size:64" json:"nickname"` // 我给对方的备注
	StartDate  time.Time  `gorm:"not null" json:"start_date"`
	Notes      string     `gorm:"type:text" json:"notes"`
	Status     string     `gorm:"size:16;default:active" json:"status"`
	EndedAt    *time.Time `json:"ended_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	User   *User `gorm:"foreignKey:UserID" json:"-"`
	Friend *User `gorm:"foreignKey:FriendID" json:"-"`
}
