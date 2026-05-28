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

		groups := make(map[string][]models.Relationship)
		for _, r := range rels {
			typeName := r.Type
			if r.CustomType != "" {
				typeName = r.CustomType
			}
			groups[typeName] = append(groups[typeName], r)
		}

		pendingCount := h.PendingCount(userID)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":        "首页",
			"userID":       userID,
			"groups":       groups,
			"rels":         rels,
			"pendingCount": pendingCount,
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

	// 检查是否有对方发来的同类型待处理申请（直接通过）
	var pendingReq models.Relationship
	if err := h.db.Where("user_id = ? AND friend_id = ? AND type = ? AND status = ?",
		friendID, userID, relType, "pending").First(&pendingReq).Error; err == nil {
		// 对方已申请，直接通过
		pendingReq.Status = "active"
		pendingReq.StartDate = startDate
		pendingReq.Notes = notes
		h.db.Save(&pendingReq)

		// 创建当前用户的关系
		rel := models.Relationship{
			UserID: uint(userID), FriendID: uint(friendID),
			Type: relType, CustomType: customType,
			Nickname: nickname, StartDate: startDate, Notes: notes,
			Status: "active",
		}
		h.db.Create(&rel)
		c.Redirect(http.StatusFound, "/friends")
		return
	}

	// 创建待处理申请
	rel := models.Relationship{
		UserID:     uint(userID),
		FriendID:   uint(friendID),
		Type:       relType,
		CustomType: customType,
		Nickname:   nickname,
		StartDate:  startDate,
		Notes:      notes,
		Status:     "pending",
	}
	h.db.Create(&rel)

	c.HTML(http.StatusOK, "friends_add.html", gin.H{
		"title":  "添加好友",
		"success": "好友申请已发送，等待对方通过",
	})
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

// FriendsRequests 查看待处理的好友申请
func (h *RelationshipHandler) FriendsRequests(c *gin.Context) {
	userID := c.GetUint("userID")

	var requests []models.Relationship
	h.db.Where("friend_id = ? AND status = ?", userID, "pending").
		Preload("User").
		Order("created_at DESC").
		Find(&requests)

	c.HTML(http.StatusOK, "friends_requests.html", gin.H{
		"title":    "好友申请",
		"requests": requests,
	})
}

// FriendAccept 通过好友申请
func (h *RelationshipHandler) FriendAccept(c *gin.Context) {
	userID := c.GetUint("userID")
	relID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req models.Relationship
	if err := h.db.Where("id = ? AND friend_id = ? AND status = ?", relID, userID, "pending").First(&req).Error; err != nil {
		c.Redirect(http.StatusFound, "/friends/requests")
		return
	}

	req.Status = "active"
	req.StartDate = time.Now()
	h.db.Save(&req)

	// 给对方创建反向关系
	var existing models.Relationship
	if err := h.db.Where("user_id = ? AND friend_id = ? AND type = ?", userID, req.UserID, req.Type).First(&existing).Error; err != nil {
		rel := models.Relationship{
			UserID: uint(userID), FriendID: req.UserID,
			Type: req.Type, CustomType: req.CustomType,
			StartDate: time.Now(), Status: "active",
		}
		h.db.Create(&rel)
	} else if existing.Status != "active" {
		existing.Status = "active"
		existing.EndedAt = nil
		h.db.Save(&existing)
	}

	c.Redirect(http.StatusFound, "/friends/requests")
}

// FriendDecline 拒绝好友申请
func (h *RelationshipHandler) FriendDecline(c *gin.Context) {
	userID := c.GetUint("userID")
	relID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req models.Relationship
	if err := h.db.Where("id = ? AND friend_id = ? AND status = ?", relID, userID, "pending").First(&req).Error; err != nil {
		c.Redirect(http.StatusFound, "/friends/requests")
		return
	}

	req.Status = "declined"
	now := time.Now()
	req.EndedAt = &now
	h.db.Save(&req)

	c.Redirect(http.StatusFound, "/friends/requests")
}

// PendingCount 返回待处理申请数
func (h *RelationshipHandler) PendingCount(userID uint) int64 {
	var count int64
	h.db.Model(&models.Relationship{}).
		Where("friend_id = ? AND status = ?", userID, "pending").
		Count(&count)
	return count
}

// Notifications 消息中心 — 好友申请 + 最近聊天
func (h *RelationshipHandler) Notifications(c *gin.Context) {
	userID := c.GetUint("userID")

	var requests []models.Relationship
	h.db.Where("friend_id = ? AND status = ?", userID, "pending").
		Preload("User").
		Order("created_at DESC").
		Find(&requests)

	type RecentChat struct {
		RelID      uint
		FriendName string
		LastMsg    string
		Time       string
	}
	var recentChats []RecentChat

	var rels []models.Relationship
	h.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Friend").Find(&rels)

	for _, r := range rels {
		var msg models.ChatMessage
		if err := h.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, r.FriendID, r.FriendID, userID).
			Order("created_at DESC").First(&msg).Error; err == nil {
			friendName := ""
			if r.Nickname != "" {
				friendName = r.Nickname
			} else if r.Friend != nil {
				friendName = r.Friend.Nickname
			}
			lastMsg := msg.Content
			if lastMsg == "" && msg.ImagePath != "" {
				lastMsg = "[图片]"
			}
			recentChats = append(recentChats, RecentChat{
				RelID:      r.ID,
				FriendName: friendName,
				LastMsg:    lastMsg,
				Time:       msg.CreatedAt.Format("01-02 15:04"),
			})
		}
	}

	c.HTML(http.StatusOK, "notifications.html", gin.H{
		"title":       "消息中心",
		"requests":    requests,
		"recentChats": recentChats,
	})
}

