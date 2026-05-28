package models

import "time"

// Comment 日记评论
type Comment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DiaryEntryID uint     `gorm:"not null;index" json:"diary_entry_id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	CreatedAt   time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user"`
}

// Like 点赞
type Like struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiaryEntryID uint      `gorm:"not null;index:idx_like_entry_user,unique" json:"diary_entry_id"`
	UserID       uint      `gorm:"not null;index:idx_like_entry_user,unique" json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}
