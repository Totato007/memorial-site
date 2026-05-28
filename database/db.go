package database

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"memorial-site/config"
	"memorial-site/models"
	"memorial-site/services"
)

// Init 初始化 SQLite 数据库：创建目录、自动迁移、种子管理员账号
func Init(cfg *config.Config, authSvc *services.AuthService) (*gorm.DB, error) {
	// 确保数据目录存在
	os.MkdirAll(filepath.Dir(cfg.DBPath), 0755)

	// 确保上传子目录存在
	os.MkdirAll(cfg.UploadDir+"/avatars", 0755)
	os.MkdirAll(cfg.UploadDir+"/chat", 0755)
	os.MkdirAll(cfg.UploadDir+"/albums", 0755)
	os.MkdirAll(cfg.UploadDir+"/diary", 0755)

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移所有模型
	db.AutoMigrate(
		&models.User{},
		&models.Relationship{},
		&models.Plan{},
		&models.ChatMessage{},
		&models.DiaryEntry{},
		&models.PhotoAlbum{},
		&models.Photo{},
		&models.Tag{},
		&models.Tagging{},
	)

	// 种子管理员账号 — 仅在不存在时创建
	var admin models.User
	if err := db.Where("username = ?", cfg.AdminUser).First(&admin).Error; err != nil {
		hash, _ := authSvc.HashPassword(cfg.AdminPass)
		db.Create(&models.User{
			Username:     cfg.AdminUser,
			Phone:        "00000000000",
			PasswordHash: hash,
			Nickname:     "管理员",
			Role:         "admin",
			IsActive:     true,
		})
	}

	return db, nil
}
