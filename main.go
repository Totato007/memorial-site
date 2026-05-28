package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"memorial-site/config"
	"memorial-site/database"
	"memorial-site/handlers"
	"memorial-site/middleware"
	"memorial-site/services"
)

func collectHTML(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func firstChar(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return ""
	}
	return string(r)
}

func main() {
	cfg := config.Load()

	authSvc := services.NewAuthService(cfg)
	imgSvc := services.NewImageService(cfg)

	db, err := database.Init(cfg, authSvc)
	if err != nil {
		panic("数据库初始化失败: " + err.Error())
	}

	r := gin.Default()
	r.MaxMultipartMemory = 4 << 20

	// 自定义模板函数
	r.SetFuncMap(template.FuncMap{
		"firstChar": firstChar,
		"onlineStatus": func(t *time.Time) string {
			s, _ := services.OnlineStatus(t)
			return s
		},
		"onlineClass": func(t *time.Time) string {
			_, c := services.OnlineStatus(t)
			return c
		},
	})
	r.LoadHTMLFiles(collectHTML("templates")...)

	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	authH := handlers.NewAuthHandler(db, authSvc)
	relH := handlers.NewRelationshipHandler(db)
	planH := handlers.NewPlanHandler(db)
	chatH := handlers.NewChatHandler(db, imgSvc)
	diaryH := handlers.NewDiaryHandler(db, imgSvc)
	albumH := handlers.NewAlbumHandler(db, imgSvc)
	adminH := handlers.NewAdminHandler(db)

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })
	r.GET("/login", authH.LoginPage)
	r.POST("/login", authH.Login)
	r.GET("/register", authH.RegisterPage)
	r.POST("/register", authH.Register)

	auth := r.Group("/")
	auth.Use(middleware.JWTAuth(authSvc, db))
	{
		auth.POST("/logout", authH.Logout)
		auth.GET("/dashboard", relH.Dashboard)
		auth.GET("/friends", relH.FriendsList)
		auth.GET("/friends/add", relH.FriendsAddPage)
		auth.POST("/friends/add", relH.FriendsAdd)
		auth.GET("/friends/:id", relH.FriendDetail)
		auth.POST("/friends/:id/edit", relH.FriendEdit)
		auth.POST("/friends/:id/remove", relH.FriendRemove)
		auth.GET("/plans", planH.PlanList)
		auth.POST("/plans", planH.PlanCreate)
		auth.POST("/plans/:id", planH.PlanUpdate)
		auth.POST("/plans/:id/status", planH.PlanStatusUpdate)
		auth.POST("/plans/:id/delete", planH.PlanDelete)
		auth.GET("/chat", chatH.ChatPage)
		auth.POST("/chat/send", chatH.ChatSend)
		auth.GET("/chat/poll", chatH.ChatPoll)
		auth.GET("/diary", diaryH.DiaryList)
		auth.POST("/diary", diaryH.DiaryCreate)
		auth.POST("/diary/:id", diaryH.DiaryUpdate)
		auth.POST("/diary/:id/delete", diaryH.DiaryDelete)
		auth.GET("/public", diaryH.PublicFeed)
		auth.GET("/albums", albumH.AlbumList)
		auth.POST("/albums", albumH.AlbumCreate)
		auth.GET("/albums/:id", albumH.AlbumDetail)
		auth.POST("/albums/:id/photos", albumH.PhotoUpload)
		auth.POST("/albums/:id/delete", albumH.AlbumDelete)
		auth.POST("/albums/:id/photos/:pid/delete", albumH.PhotoDelete)
		auth.GET("/profile", authH.ProfilePage)
		auth.POST("/profile", authH.ProfileUpdate)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth(authSvc, db), middleware.RequireAdmin())
	{
		admin.GET("/", adminH.AdminDashboard)
		admin.GET("/users", adminH.AdminUsers)
		admin.POST("/users/:id/toggle", adminH.AdminToggleUser)
	}

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
