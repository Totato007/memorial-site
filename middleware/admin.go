package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin 管理员角色校验 — 需在 JWTAuth 之后使用
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.String(http.StatusForbidden, "无权访问")
			c.Abort()
			return
		}
		c.Next()
	}
}
