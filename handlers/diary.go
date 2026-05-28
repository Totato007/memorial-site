package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
	"memorial-site/services"
)

type DiaryHandler struct {
	db     *gorm.DB
	imgSvc *services.ImageService
}

func NewDiaryHandler(db *gorm.DB, imgSvc *services.ImageService) *DiaryHandler {
	return &DiaryHandler{db: db, imgSvc: imgSvc}
}

// DiaryEntryWithMeta 日记 + 点赞数 + 评论
type DiaryEntryWithMeta struct {
	models.DiaryEntry
	LikeCount    int64            `json:"like_count"`
	UserLiked    bool             `json:"user_liked"`
	Comments     []models.Comment `json:"comments"`
	SmartTime    string           `json:"smart_time"`
	SmartDate    string           `json:"smart_date"`
}

// DiaryList 我的日记列表
func (h *DiaryHandler) DiaryList(c *gin.Context) {
	userID := c.GetUint("userID")

	var entries []models.DiaryEntry
	query := h.db.Where("user_id = ?", userID).Order("entry_date DESC, created_at DESC")
	if month := c.Query("month"); month != "" {
		if t, err := time.Parse("2006-01", month); err == nil {
			start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
			end := start.AddDate(0, 1, 0)
			query = query.Where("entry_date >= ? AND entry_date < ?", start, end)
		}
	}
	query.Limit(30).Find(&entries)

	entriesWithMeta := h.enrichEntries(entries, userID)

	c.HTML(http.StatusOK, "diary.html", gin.H{
		"title":    "心情日记",
		"entries":  entriesWithMeta,
		"month":    c.Query("month"),
		"userID":   userID,
	})
}

// PublicFeed 公共广场
func (h *DiaryHandler) PublicFeed(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := uint(0)
	if userID != nil {
		uid = userID.(uint)
	}

	var entries []models.DiaryEntry
	h.db.Where("visibility = ?", "public").
		Preload("User").
		Order("created_at DESC").
		Limit(30).
		Find(&entries)

	entriesWithMeta := h.enrichEntries(entries, uid)

	c.HTML(http.StatusOK, "public_feed.html", gin.H{
		"title":   "公共广场",
		"entries": entriesWithMeta,
		"userID":  uid,
	})
}

// DiaryCreate 创建日记
func (h *DiaryHandler) DiaryCreate(c *gin.Context) {
	userID := c.GetUint("userID")
	content := c.PostForm("content")
	mood := c.PostForm("mood")
	visibility := c.PostForm("visibility")
	entryDateStr := c.PostForm("entry_date")

	if content == "" {
		c.Redirect(http.StatusFound, "/diary")
		return
	}

	entryDate := time.Now()
	if entryDateStr != "" {
		if t, err := time.Parse("2006-01-02", entryDateStr); err == nil {
			entryDate = t
		}
	}
	if mood == "" {
		mood = "neutral"
	}
	if visibility == "" {
		visibility = "friends"
	}

	entry := models.DiaryEntry{
		UserID: userID, Mood: mood, Content: content,
		Visibility: visibility, EntryDate: entryDate,
	}

	if file, err := c.FormFile("image"); err == nil {
		if file.Size <= 2<<20 && services.AllowedImageExt(file.Filename) {
			src, _ := file.Open()
			defer src.Close()
			ext := filepath.Ext(file.Filename)
			if origPath, _, err := h.imgSvc.ProcessImage(src, ext, "diary"); err == nil {
				entry.ImagePath = origPath
			}
		}
	}
	h.db.Create(&entry)
	c.Redirect(http.StatusFound, "/diary")
}

// DiaryUpdate 更新日记
func (h *DiaryHandler) DiaryUpdate(c *gin.Context) {
	userID := c.GetUint("userID")
	entryID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var entry models.DiaryEntry
	if err := h.db.Where("id = ? AND user_id = ?", entryID, userID).First(&entry).Error; err != nil {
		c.Redirect(http.StatusFound, "/diary")
		return
	}

	entry.Content = c.PostForm("content")
	entry.Mood = c.PostForm("mood")
	if v := c.PostForm("visibility"); v != "" {
		entry.Visibility = v
	}
	if d := c.PostForm("entry_date"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			entry.EntryDate = t
		}
	}
	h.db.Save(&entry)
	c.Redirect(http.StatusFound, "/diary")
}

// DiaryDelete 删除日记
func (h *DiaryHandler) DiaryDelete(c *gin.Context) {
	userID := c.GetUint("userID")
	entryID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Where("id = ? AND user_id = ?", entryID, userID).Delete(&models.DiaryEntry{})
	c.Redirect(http.StatusFound, "/diary")
}

// CommentCreate 添加评论 (AJAX)
func (h *DiaryHandler) CommentCreate(c *gin.Context) {
	userID := c.GetUint("userID")
	entryID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	content := c.PostForm("content")
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论不能为空"})
		return
	}

	comment := models.Comment{
		DiaryEntryID: uint(entryID),
		UserID:       userID,
		Content:      content,
	}
	h.db.Create(&comment)
	h.db.Preload("User").First(&comment, comment.ID)

	c.JSON(http.StatusOK, comment)
}

// LikeToggle 切换点赞 (AJAX)
func (h *DiaryHandler) LikeToggle(c *gin.Context) {
	userID := c.GetUint("userID")
	entryID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var existing models.Like
	if err := h.db.Where("diary_entry_id = ? AND user_id = ?", entryID, userID).First(&existing).Error; err == nil {
		h.db.Delete(&existing)
	} else {
		h.db.Create(&models.Like{DiaryEntryID: uint(entryID), UserID: userID})
	}

	var count int64
	h.db.Model(&models.Like{}).Where("diary_entry_id = ?", entryID).Count(&count)

	liked := h.db.Where("diary_entry_id = ? AND user_id = ?", entryID, userID).First(&models.Like{}).Error == nil

	c.JSON(http.StatusOK, gin.H{"liked": liked, "count": count})
}

// enrichEntries 给日记列表附加上点赞数、评论、时间显示
func (h *DiaryHandler) enrichEntries(entries []models.DiaryEntry, userID uint) []DiaryEntryWithMeta {
	var result []DiaryEntryWithMeta
	for _, e := range entries {
		var likeCount int64
		h.db.Model(&models.Like{}).Where("diary_entry_id = ?", e.ID).Count(&likeCount)

		userLiked := false
		if userID > 0 {
			var l models.Like
			userLiked = h.db.Where("diary_entry_id = ? AND user_id = ?", e.ID, userID).First(&l).Error == nil
		}

		var comments []models.Comment
		h.db.Where("diary_entry_id = ?", e.ID).Preload("User").Order("created_at ASC").Limit(20).Find(&comments)

		result = append(result, DiaryEntryWithMeta{
			DiaryEntry: e,
			LikeCount:  likeCount,
			UserLiked:  userLiked,
			Comments:   comments,
			SmartTime:  services.SmartTime(e.CreatedAt),
			SmartDate:  services.SmartDate(e.EntryDate),
		})
	}
	return result
}
