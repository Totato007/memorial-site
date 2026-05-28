package models

import "time"

// PhotoAlbum 专属相册
type PhotoAlbum struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	RelationshipID uint      `gorm:"not null;index" json:"relationship_id"`
	CreatorID      uint      `gorm:"not null" json:"creator_id"`
	Title          string    `gorm:"size:128;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	CoverImage     string    `gorm:"size:255" json:"cover_image"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Creator *User   `gorm:"foreignKey:CreatorID" json:"-"`
	Photos  []Photo `gorm:"foreignKey:AlbumID" json:"-"`
}

// Photo 相册内的单张照片
type Photo struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AlbumID       uint      `gorm:"not null;index" json:"album_id"`
	UploaderID    uint      `gorm:"not null" json:"uploader_id"`
	ImagePath     string    `gorm:"size:255;not null" json:"image_path"`
	ThumbnailPath string    `gorm:"size:255" json:"thumbnail_path"`
	Description   string    `gorm:"type:text" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Tag 通用标签
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Color     string    `gorm:"size:7;default:#6366f1" json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// Tagging 多态标签关联
type Tagging struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TagID      uint   `gorm:"not null;index" json:"tag_id"`
	TaggedType string `gorm:"size:32;not null;index:idx_taggings" json:"tagged_type"` // plan | diary | album
	TaggedID   uint   `gorm:"not null;index:idx_taggings" json:"tagged_id"`
}
