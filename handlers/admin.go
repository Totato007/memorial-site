package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// AdminDashboard 管理后台首页 — 站点概览数据
func (h *AdminHandler) AdminDashboard(c *gin.Context) {
	var userCount, relCount, photoCount, msgCount int64
	h.db.Model(&models.User{}).Count(&userCount)
	h.db.Model(&models.Relationship{}).Where("status = ?", "active").Count(&relCount)
	h.db.Model(&models.Photo{}).Count(&photoCount)
	h.db.Model(&models.ChatMessage{}).Count(&msgCount)

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title":      "管理后台",
		"userCount":  userCount,
		"relCount":   relCount,
		"photoCount": photoCount,
		"msgCount":   msgCount,
	})
}

// AdminUsers 用户管理列表
func (h *AdminHandler) AdminUsers(c *gin.Context) {
	var users []models.User
	h.db.Order("created_at DESC").Find(&users)

	c.HTML(http.StatusOK, "admin_users.html", gin.H{
		"title": "用户管理",
		"users": users,
	})
}

// AdminToggleUser 启用/禁用用户
func (h *AdminHandler) AdminToggleUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 不能禁用自己
	if adminID := c.GetUint("userID"); adminID == uint(userID) {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	user.IsActive = !user.IsActive
	h.db.Save(&user)
	c.Redirect(http.StatusFound, "/admin/users")
}
