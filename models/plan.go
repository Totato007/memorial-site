package models

import "time"

// Plan 未来计划/待办事项
type Plan struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	RelationshipID uint       `gorm:"not null;index" json:"relationship_id"`
	CreatorID      uint       `gorm:"not null" json:"creator_id"`
	Title          string     `gorm:"size:128;not null" json:"title"`
	Content        string     `gorm:"type:text" json:"content"`
	Category       string     `gorm:"size:64;default:general" json:"category"`
	Status         string     `gorm:"size:16;default:not_started" json:"status"`
	Visibility     string     `gorm:"size:16;default:friends" json:"visibility"` // private|couple|friends|public
	TargetDate     *time.Time `json:"target_date"`
	SortOrder      int        `gorm:"default:0" json:"sort_order"`
	IsArchived     bool       `gorm:"default:false" json:"is_archived"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatorID" json:"-"`
}
