package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
	"memorial-site/services"
)

// JWTAuth JWT 鉴权中间件 — 未登录重定向到 /login，同时更新在线状态
func JWTAuth(authSvc *services.AuthService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		claims, err := authSvc.ParseToken(tokenStr)
		if err != nil {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 更新最后活跃时间 (在线状态判定用)
		now := time.Now()
		db.Model(&models.User{}).Where("id = ?", claims.UserID).Update("last_active_at", &now)

		c.Next()
	}
}

// SetUserIfExists 尝试解析 token 但不强制登录 — 用于公开页面
func SetUserIfExists(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("token")
		if err != nil {
			c.Next()
			return
		}
		claims, err := authSvc.ParseToken(tokenStr)
		if err != nil {
			c.Next()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
