package services

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // 注册 PNG 解码器
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"

	"memorial-site/config"
)

type ImageService struct {
	cfg *config.Config
}

func NewImageService(cfg *config.Config) *ImageService {
	return &ImageService{cfg: cfg}
}

// ProcessImage 处理上传图片: 校验格式 → 解码 → 缩放 → JPEG压缩 → 生成缩略图
// 返回: 压缩后原图路径, 缩略图路径, error
// 配置项: config.MaxImgSize, config.ImgMaxWidth, config.ImgQuality
func (s *ImageService) ProcessImage(file io.Reader, ext string, subDir string) (origPath string, thumbPath string, err error) {
	// 校验文件后缀，仅允许 jpeg/png/webp
	ext = strings.ToLower(ext)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", "", fmt.Errorf("不支持的图片格式: %s", ext)
	}

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("图片解码失败: %v", err)
	}

	// 缩放 — 宽度超过最大值则等比缩小
	bounds := img.Bounds()
	if bounds.Dx() > s.cfg.ImgMaxWidth {
		newHeight := bounds.Dy() * s.cfg.ImgMaxWidth / bounds.Dx()
		img = resizeImage(img, s.cfg.ImgMaxWidth, newHeight)
	}

	// 生成文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dir := filepath.Join(s.cfg.UploadDir, subDir)
	os.MkdirAll(dir, 0755)

	// 保存压缩后的原图 (统一输出 JPEG)
	origFilename := filename
	if format != "jpeg" {
		origFilename = strings.TrimSuffix(filename, ext) + ".jpg"
	}
	origPath = filepath.Join(subDir, origFilename)
	fullOrigPath := filepath.Join(s.cfg.UploadDir, origPath)
	outFile, err := os.Create(fullOrigPath)
	if err != nil {
		return "", "", err
	}
	defer outFile.Close()
	if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: s.cfg.ImgQuality}); err != nil {
		return "", "", err
	}

	// 生成缩略图 (200px 宽)
	thumbImg := resizeImage(img, 200, 200*bounds.Dy()/bounds.Dx())
	if thumbImg.Bounds().Dy() < 1 {
		return "", "", err
	}
	thumbFilename := "thumb_" + origFilename
	thumbPath = filepath.Join(subDir, thumbFilename)
	fullThumbPath := filepath.Join(s.cfg.UploadDir, thumbPath)
	thumbFile, err := os.Create(fullThumbPath)
	if err != nil {
		return "", "", err
	}
	defer thumbFile.Close()
	jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 70})

	return origPath, thumbPath, nil
}

// resizeImage 等比缩放图片到指定宽高
func resizeImage(img image.Image, width, height int) image.Image {
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// DetectMIMEType 从文件名检测 MIME 类型
func DetectMIMEType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

// AllowedImageExt 检查是否为允许的图片后缀
func AllowedImageExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
