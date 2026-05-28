package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
	"memorial-site/services"
)

type RelationshipHandler struct {
	db *gorm.DB
}

func NewRelationshipHandler(db *gorm.DB) *RelationshipHandler {
	return &RelationshipHandler{db: db}
}

// Dashboard 首页 — 好友列表 + 快捷入口
func (h *RelationshipHandler) Dashboard(c *gin.Context) {
	userID := c.GetUint("userID")

	var currentUser models.User
	h.db.First(&currentUser, userID)

	var rels []models.Relationship
	h.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Friend").
		Order("type ASC, created_at DESC").
		Find(&rels)

	// 按类型分组
	groups := make(map[string][]models.Relationship)
	for _, r := range rels {
		typeName := r.Type
		if r.CustomType != "" {
			typeName = r.CustomType
		}
		groups[typeName] = append(groups[typeName], r)
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":  "首页",
		"userID": userID,
		"groups": groups,
		"rels":   rels,
	})
}

// FriendsList 好友列表页
func (h *RelationshipHandler) FriendsList(c *gin.Context) {
	userID := c.GetUint("userID")

	var rels []models.Relationship
	h.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Friend").
		Order("type ASC, created_at DESC").
		Find(&rels)

	groups := make(map[string][]models.Relationship)
	for _, r := range rels {
		typeName := r.Type
		if r.CustomType != "" {
			typeName = r.CustomType
		}
		groups[typeName] = append(groups[typeName], r)
	}

	c.HTML(http.StatusOK, "friends.html", gin.H{
		"title":  "好友列表",
		"groups": groups,
		"rels":   rels,
	})
}

// FriendsAddPage 添加好友页
func (h *RelationshipHandler) FriendsAddPage(c *gin.Context) {
	c.HTML(http.StatusOK, "friends_add.html", gin.H{
		"title": "添加好友",
	})
}

// FriendsAdd 搜索用户并添加好友
func (h *RelationshipHandler) FriendsAdd(c *gin.Context) {
	userID := c.GetUint("userID")
	query := c.PostForm("query")
	relType := c.PostForm("type")
	customType := c.PostForm("custom_type")

	if query != "" {
		// 搜索用户
		var users []models.User
		h.db.Where("(username = ? OR phone = ? OR nickname LIKE ?) AND id != ? AND is_active = ?",
			query, query, "%"+query+"%", userID, true).
			Limit(10).Find(&users)

		c.HTML(http.StatusOK, "friends_add.html", gin.H{
			"title":      "添加好友",
			"searchResults": users,
			"searchQuery":   query,
			"relType":       relType,
			"customType":    customType,
		})
		return
	}

	// 确认添加
	friendIDStr := c.PostForm("friend_id")
	startDateStr := c.PostForm("start_date")
	notes := c.PostForm("notes")
	nickname := c.PostForm("nickname")

	if friendIDStr == "" {
		c.HTML(http.StatusOK, "friends_add.html", gin.H{
			"title": "添加好友",
			"error": "请先搜索并选择用户",
		})
		return
	}

	friendID, _ := strconv.ParseUint(friendIDStr, 10, 64)

	if relType == "" {
		relType = "friend"
	}

	// 情侣类型约束: 每人最多1个
	if relType == "couple" {
		var existing models.Relationship
		if err := h.db.Where("user_id = ? AND type = ? AND status = ?", userID, "couple", "active").First(&existing).Error; err == nil {
			c.HTML(http.StatusOK, "friends_add.html", gin.H{
				"title": "添加好友",
				"error": "你已有一个情侣关系，不能同时拥有多个情侣",
			})
			return
		}
	}

	// 检查是否已存在同类型活跃关系
	var dup models.Relationship
	if err := h.db.Where("user_id = ? AND friend_id = ? AND type = ? AND status = ?", userID, friendID, relType, "active").First(&dup).Error; err == nil {
		c.HTML(http.StatusOK, "friends_add.html", gin.H{
			"title": "添加好友",
			"error": "你们已经是这种关系了",
		})
		return
	}

	startDate := time.Now()
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}

	// 创建双向关系
	rel1 := models.Relationship{
		UserID:     uint(userID),
		FriendID:   uint(friendID),
		Type:       relType,
		CustomType: customType,
		Nickname:   nickname,
		StartDate:  startDate,
		Notes:      notes,
	}
	h.db.Create(&rel1)

	// 给对方也创建一条 (反向关系)
	var friendUser models.User
	if err := h.db.First(&friendUser, friendID).Error; err == nil {
		friendNickname := c.PostForm("friend_nickname")
		rel2 := models.Relationship{
			UserID:     uint(friendID),
			FriendID:   uint(userID),
			Type:       relType,
			CustomType: customType,
			Nickname:   friendNickname,
			StartDate:  startDate,
			Notes:      notes,
		}
		h.db.Create(&rel2)
	}

	c.Redirect(http.StatusFound, "/friends")
}

// FriendDetail 好友详情
func (h *RelationshipHandler) FriendDetail(c *gin.Context) {
	userID := c.GetUint("userID")
	relID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var rel models.Relationship
	if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").
		Preload("Friend").First(&rel).Error; err != nil {
		c.Redirect(http.StatusFound, "/friends")
		return
	}

	daysTogether := services.CalculateDaysTogether(rel.StartDate)
	_, daysUntil := services.NextAnniversary(rel.StartDate)

	c.HTML(http.StatusOK, "friend_detail.html", gin.H{
		"title":        rel.Friend.Nickname,
		"rel":          rel,
		"daysTogether": daysTogether,
		"daysUntil":    daysUntil,
		"startDateFmt": services.FormatDate(rel.StartDate),
	})
}

// FriendEdit 编辑好友关系
func (h *RelationshipHandler) FriendEdit(c *gin.Context) {
	userID := c.GetUint("userID")
	relID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var rel models.Relationship
	if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").First(&rel).Error; err != nil {
		c.Redirect(http.StatusFound, "/friends")
		return
	}

	rel.Nickname = c.PostForm("nickname")
	rel.Notes = c.PostForm("notes")
	if customType := c.PostForm("custom_type"); customType != "" {
		rel.CustomType = customType
	}
	if startDateStr := c.PostForm("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			rel.StartDate = t
		}
	}

	h.db.Save(&rel)
	c.Redirect(http.StatusFound, "/friends/"+c.Param("id"))
}

// FriendRemove 删除好友
func (h *RelationshipHandler) FriendRemove(c *gin.Context) {
	userID := c.GetUint("userID")
	relID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var rel models.Relationship
	if err := h.db.Where("id = ? AND user_id = ? AND status = ?", relID, userID, "active").First(&rel).Error; err != nil {
		c.Redirect(http.StatusFound, "/friends")
		return
	}

	now := time.Now()
	rel.Status = "ended"
	rel.EndedAt = &now
	h.db.Save(&rel)

	// 同时结束反向关系
	h.db.Model(&models.Relationship{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", rel.FriendID, rel.UserID, "active").
		Updates(map[string]interface{}{"status": "ended", "ended_at": &now})

	c.Redirect(http.StatusFound, "/friends")
}
