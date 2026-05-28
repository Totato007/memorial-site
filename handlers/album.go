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

type AlbumHandler struct {
	db     *gorm.DB
	imgSvc *services.ImageService
}

func NewAlbumHandler(db *gorm.DB, imgSvc *services.ImageService) *AlbumHandler {
	return &AlbumHandler{db: db, imgSvc: imgSvc}
}

// AlbumList 相册列表
func (h *AlbumHandler) AlbumList(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	var albums []models.PhotoAlbum
	h.db.Where("relationship_id = ?", rel.ID).Order("created_at DESC").Find(&albums)

	c.HTML(http.StatusOK, "album_list.html", gin.H{
		"title":  "专属相册",
		"albums": albums,
	})
}

// AlbumCreate 创建相册
func (h *AlbumHandler) AlbumCreate(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		c.Redirect(http.StatusFound, "/albums")
		return
	}

	album := models.PhotoAlbum{
		RelationshipID: rel.ID,
		CreatorID:      userID,
		Title:          title,
		Description:    c.PostForm("description"),
	}
	h.db.Create(&album)

	c.Redirect(http.StatusFound, "/albums")
}

// AlbumDetail 查看相册中的照片
func (h *AlbumHandler) AlbumDetail(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var album models.PhotoAlbum
	if err := h.db.Where("id = ? AND relationship_id = ?", albumID, rel.ID).First(&album).Error; err != nil {
		c.Redirect(http.StatusFound, "/albums")
		return
	}

	var photos []models.Photo
	h.db.Where("album_id = ?", album.ID).Order("created_at DESC").Find(&photos)

	c.HTML(http.StatusOK, "album_detail.html", gin.H{
		"title":  album.Title,
		"album":  album,
		"photos": photos,
	})
}

// PhotoUpload 上传照片到相册
func (h *AlbumHandler) PhotoUpload(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var album models.PhotoAlbum
	if err := h.db.Where("id = ? AND relationship_id = ?", albumID, rel.ID).First(&album).Error; err != nil {
		c.Redirect(http.StatusFound, "/albums")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.Redirect(http.StatusFound, "/albums/"+c.Param("id"))
		return
	}

	if file.Size > 2<<20 || !services.AllowedImageExt(file.Filename) {
		c.Redirect(http.StatusFound, "/albums/"+c.Param("id"))
		return
	}

	ext := filepath.Ext(file.Filename)
	src, _ := file.Open()
	defer src.Close()

	origPath, thumbPath, err := h.imgSvc.ProcessImage(src, ext, "albums")
	if err != nil {
		c.Redirect(http.StatusFound, "/albums/"+c.Param("id"))
		return
	}

	photo := models.Photo{
		AlbumID:       uint(albumID),
		UploaderID:    userID,
		ImagePath:     origPath,
		ThumbnailPath: thumbPath,
		Description:   c.PostForm("description"),
	}
	h.db.Create(&photo)

	// 设置封面 (使用第一张照片)
	if album.CoverImage == "" {
		album.CoverImage = thumbPath
		h.db.Save(&album)
	}

	c.Redirect(http.StatusFound, "/albums/"+c.Param("id"))
}

// AlbumDelete 删除相册
func (h *AlbumHandler) AlbumDelete(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Where("id = ? AND relationship_id = ?", albumID, rel.ID).Delete(&models.Photo{})
	h.db.Where("id = ? AND relationship_id = ?", albumID, rel.ID).Delete(&models.PhotoAlbum{})
	c.Redirect(http.StatusFound, "/albums")
}

// PhotoDelete 删除单张照片
func (h *AlbumHandler) PhotoDelete(c *gin.Context) {
	userID := c.GetUint("userID")
	rel := h.findUserRelationship(userID)
	if rel == nil {
		c.HTML(http.StatusOK, "album_list.html", gin.H{
			"title":  "专属相册",
			"albums": nil,
			"noRel":  true,
		})
		return
	}

	photoID, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	h.db.Where("id = ? AND album_id IN (SELECT id FROM photo_albums WHERE relationship_id = ?)", photoID, rel.ID).Delete(&models.Photo{})
	c.Redirect(http.StatusFound, "/albums/"+c.Param("id"))
}

func (h *AlbumHandler) findUserRelationship(userID uint) *models.Relationship {
	var rel models.Relationship
	err := h.db.Where("user_id = ? AND status = ?", userID, "active").First(&rel).Error
	if err != nil {
		return nil
	}
	return &rel
}
