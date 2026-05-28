package models

import "time"

// DiaryEntry 心情日记 — 支持四级可见权限
type DiaryEntry struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Mood       string    `gorm:"size:16;default:neutral" json:"mood"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	ImagePath  string    `gorm:"size:255" json:"image_path"`
	Visibility string    `gorm:"size:16;default:friends" json:"visibility"` // private|couple|friends|public
	EntryDate  time.Time `gorm:"not null;index:idx_diary_user_date" json:"entry_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}
