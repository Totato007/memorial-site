package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"memorial-site/models"
)

type PlanHandler struct {
	db *gorm.DB
}

func NewPlanHandler(db *gorm.DB) *PlanHandler {
	return &PlanHandler{db: db}
}

// PlanList 我的计划列表
func (h *PlanHandler) PlanList(c *gin.Context) {
	userID := c.GetUint("userID")

	statusFilter := c.Query("status")
	categoryFilter := c.Query("category")

	query := h.db.Where("creator_id = ? AND is_archived = ?", userID, false)
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if categoryFilter != "" {
		query = query.Where("category = ?", categoryFilter)
	}

	var plans []models.Plan
	query.Order("target_date ASC, created_at DESC").Find(&plans)

	c.HTML(http.StatusOK, "plans.html", gin.H{
		"title":    "未来计划",
		"plans":    plans,
		"status":   statusFilter,
		"category": categoryFilter,
	})
}

// PlanCreate 创建计划
func (h *PlanHandler) PlanCreate(c *gin.Context) {
	userID := c.GetUint("userID")

	title := c.PostForm("title")
	if title == "" {
		c.HTML(http.StatusOK, "plans.html", gin.H{"title": "未来计划", "error": "计划名称不能为空"})
		return
	}

	category := c.PostForm("category")
	if category == "" {
		category = "general"
	}
	visibility := c.PostForm("visibility")
	if visibility == "" {
		visibility = "friends"
	}

	var targetDate *time.Time
	if d := c.PostForm("target_date"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			targetDate = &t
		}
	}

	plan := models.Plan{
		CreatorID:  userID,
		Title:      title,
		Content:    c.PostForm("content"),
		Category:   category,
		Visibility: visibility,
		Status:     "not_started",
		TargetDate: targetDate,
	}
	h.db.Create(&plan)
	c.Redirect(http.StatusFound, "/plans")
}

// PlanUpdate 更新计划
func (h *PlanHandler) PlanUpdate(c *gin.Context) {
	userID := c.GetUint("userID")
	planID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var plan models.Plan
	if err := h.db.Where("id = ? AND creator_id = ?", planID, userID).First(&plan).Error; err != nil {
		c.Redirect(http.StatusFound, "/plans")
		return
	}

	plan.Title = c.PostForm("title")
	plan.Content = c.PostForm("content")
	plan.Category = c.PostForm("category")
	if v := c.PostForm("visibility"); v != "" {
		plan.Visibility = v
	}
	if s := c.PostForm("status"); s != "" {
		plan.Status = s
	}
	if d := c.PostForm("target_date"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			plan.TargetDate = &t
		}
	} else {
		plan.TargetDate = nil
	}

	h.db.Save(&plan)
	c.Redirect(http.StatusFound, "/plans")
}

// PlanStatusUpdate AJAX 快速切换状态
func (h *PlanHandler) PlanStatusUpdate(c *gin.Context) {
	userID := c.GetUint("userID")
	planID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var plan models.Plan
	if err := h.db.Where("id = ? AND creator_id = ?", planID, userID).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	plan.Status = c.PostForm("status")
	h.db.Save(&plan)
	c.JSON(http.StatusOK, gin.H{"status": plan.Status})
}

// PlanDelete 删除计划
func (h *PlanHandler) PlanDelete(c *gin.Context) {
	userID := c.GetUint("userID")
	planID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Where("id = ? AND creator_id = ?", planID, userID).Delete(&models.Plan{})
	c.Redirect(http.StatusFound, "/plans")
}
