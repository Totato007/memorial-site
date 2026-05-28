package models

import "time"

// ChatMessage 聊天消息 — 文字和图片至少填一项
type ChatMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	RelationshipID uint      `gorm:"not null;index" json:"relationship_id"`
	SenderID       uint      `gorm:"not null" json:"sender_id"`
	Content        string    `gorm:"type:text" json:"content"`
	ImagePath      string    `gorm:"size:255" json:"image_path"`
	CreatedAt      time.Time `json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"-"`
}
