package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
	"memorial-site/services"
)

type ChatHandler struct {
	db     *gorm.DB
	imgSvc *services.ImageService
}

func NewChatHandler(db *gorm.DB, imgSvc *services.ImageService) *ChatHandler {
	return &ChatHandler{db: db, imgSvc: imgSvc}
}

// ChatPage 聊天页面 — 选择好友后显示对话
func (h *ChatHandler) ChatPage(c *gin.Context) {
	userID := c.GetUint("userID")

	// 加载好友列表 (用于选择聊天对象)
	var friends []models.Relationship
	h.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Friend").
		Order("type ASC").
		Find(&friends)

	// 默认选中第一个好友或通过 query 指定
	selectedID := c.Query("with")
	var messages []models.ChatMessage
	var selectedFriend *models.User
	var selectedRelID uint

	if selectedID != "" {
		relID, _ := strconv.ParseUint(selectedID, 10, 64)
		var rel models.Relationship
		if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").
			Preload("Friend").First(&rel).Error; err == nil {
			selectedFriend = rel.Friend
			selectedRelID = rel.ID

			// 加载该关系的聊天记录
			h.db.Where("relationship_id = ?", rel.ID).
				Preload("Sender").
				Order("created_at DESC").Limit(50).
				Find(&messages)
			// 倒序
			for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}

	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title":          "互动留言",
		"friends":        friends,
		"selectedRelID":  selectedRelID,
		"selectedFriend": selectedFriend,
		"messages":       messages,
		"userID":         userID,
	})
}

// ChatSend 发送消息
func (h *ChatHandler) ChatSend(c *gin.Context) {
	userID := c.GetUint("userID")
	relIDStr := c.PostForm("rel_id")
	relID, _ := strconv.ParseUint(relIDStr, 10, 64)

	// 验证关系
	var rel models.Relationship
	if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").First(&rel).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关系不存在"})
		return
	}

	content := c.PostForm("content")
	var imagePath string

	if file, err := c.FormFile("image"); err == nil {
		if file.Size > 2<<20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过2MB"})
			return
		}
		if !services.AllowedImageExt(file.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPG/PNG/WebP 格式"})
			return
		}
		src, _ := file.Open()
		defer src.Close()
		ext := filepath.Ext(file.Filename)
		if origPath, _, err := h.imgSvc.ProcessImage(src, ext, "chat"); err == nil {
			imagePath = origPath
		}
	}

	if content == "" && imagePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	msg := models.ChatMessage{
		RelationshipID: uint(relID),
		SenderID:       userID,
		Content:        content,
		ImagePath:      imagePath,
	}
	h.db.Create(&msg)

	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		h.db.Preload("Sender").First(&msg, msg.ID)
		c.JSON(http.StatusOK, msg)
		return
	}
	c.Redirect(http.StatusFound, "/chat?with="+relIDStr)
}

// ChatPoll JSON 轮询新消息
func (h *ChatHandler) ChatPoll(c *gin.Context) {
	userID := c.GetUint("userID")
	relIDStr := c.Query("rel_id")
	relID, _ := strconv.ParseUint(relIDStr, 10, 64)

	// 验证关系
	var rel models.Relationship
	if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").First(&rel).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关系不存在"})
		return
	}

	lastID, _ := strconv.ParseInt(c.Query("last_id"), 10, 64)

	var messages []models.ChatMessage
	h.db.Where("relationship_id = ? AND id > ?", rel.ID, lastID).
		Preload("Sender").Order("id ASC").Find(&messages)

	c.JSON(http.StatusOK, messages)
}
