package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
	"memorial-site/services"
)

type AuthHandler struct {
	db      *gorm.DB
	authSvc *services.AuthService
}

func NewAuthHandler(db *gorm.DB, authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{db: db, authSvc: authSvc}
}

// LoginPage 登录页面
func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "登录"})
}

// Login 处理登录表单提交
func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	if err := h.db.Where("username = ? OR phone = ?", username, username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "登录", "error": "账号或密码错误"})
		return
	}

	if !user.IsActive {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "登录", "error": "账号已被禁用，请联系管理员"})
		return
	}

	if !h.authSvc.CheckPassword(password, user.PasswordHash) {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "登录", "error": "账号或密码错误"})
		return
	}

	token, _ := h.authSvc.GenerateToken(user.ID, user.Username, user.Role)
	c.SetCookie("token", token, 72*3600, "/", "", false, true)

	if user.Role == "admin" {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	c.Redirect(http.StatusFound, "/dashboard")
}

// RegisterPage 注册页面
func (h *AuthHandler) RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{"title": "注册"})
}

// Register 处理注册表单提交
func (h *AuthHandler) Register(c *gin.Context) {
	username := c.PostForm("username")
	phone := c.PostForm("phone")
	password := c.PostForm("password")
	nickname := c.PostForm("nickname")

	if username == "" || phone == "" || password == "" {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "注册", "error": "请填写所有必填字段"})
		return
	}

	if len(password) < 6 {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "注册", "error": "密码长度不能少于6位"})
		return
	}

	// 检查用户名或手机号是否已被注册
	var existUser models.User
	if err := h.db.Where("username = ? OR phone = ?", username, phone).First(&existUser).Error; err == nil {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "注册", "error": "用户名或手机号已被注册"})
		return
	}

	hash, _ := h.authSvc.HashPassword(password)
	if nickname == "" {
		nickname = username
	}

	user := models.User{
		Username:     username,
		Phone:        phone,
		PasswordHash: hash,
		Nickname:     nickname,
		Role:         "user",
		IsActive:     true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "注册", "error": "注册失败，请稍后重试"})
		return
	}

	token, _ := h.authSvc.GenerateToken(user.ID, user.Username, user.Role)
	c.SetCookie("token", token, 72*3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/dashboard")
}

// ProfilePage 个人资料页
func (h *AuthHandler) ProfilePage(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title": "个人资料",
		"user":  user,
	})
}

// ProfileUpdate 更新个人资料
func (h *AuthHandler) ProfileUpdate(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user.Nickname = c.PostForm("nickname")
	user.Bio = c.PostForm("bio")

	if password := c.PostForm("password"); password != "" {
		if len(password) >= 6 {
			hash, _ := h.authSvc.HashPassword(password)
			user.PasswordHash = hash
		}
	}

	h.db.Save(&user)

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":   "个人资料",
		"user":    user,
		"success": "资料已更新",
	})
}

// Logout 登出 — 清除 JWT cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
