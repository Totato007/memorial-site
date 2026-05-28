package models

import "time"

// ChatMessage 聊天消息 — 基于收发双方ID
type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint      `gorm:"not null;index" json:"receiver_id"`
	Content    string    `gorm:"type:text" json:"content"`
	ImagePath  string    `gorm:"size:255" json:"image_path"`
	CreatedAt  time.Time `json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"-"`
}
