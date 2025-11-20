package routes

import (
	"newsclip/backend/internal/app/controllers"
	"newsclip/backend/internal/app/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.Static("/v1/uploads", "./uploads")
	router.Static("/v1/images", "./static/images")

	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/social", controllers.SocialLogin)
			auth.POST("/refresh", controllers.RefreshToken)
			auth.POST("/check-username", controllers.CheckUsername)

			// 그룹에 미들웨어 적용
			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}

		news := v1.Group("/news")
		{
			news.GET("/", controllers.GetNewsList)
			news.GET("/:newsId", controllers.GetNewsDetail)

			news.POST("/:newsId/interact", middlewares.AuthMiddleware(), controllers.InteractNews)
			news.POST("/:newsId/bookmark", middlewares.AuthMiddleware(), controllers.BookmarkNews)
		}

		shorts := v1.Group("/shorts")
		{
			// 1. 피드 조회 (비로그인 가능 - Optional 미들웨어)
			shorts.GET("/", middlewares.AuthMiddlewareOptional(), controllers.GetShortsFeed)
			// 2. [신규] 상호작용 (로그인 필수 - AuthMiddleware)
			shorts.POST("/:shortId/interact", middlewares.AuthMiddleware(), controllers.InteractShort)
		}

		// me 그룹에 인증 미들웨어를 붙임
		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			me.GET("/", controllers.GetMyProfile)
			me.POST("/avatar", controllers.UpdateProfile) // 필요하면 UpdateAvatar로 이름 맞추세요
			me.GET("/bookmarks", controllers.GetMyBookmarks)
		}

		community := v1.Group("/community")
		{
			community.GET("/posts", controllers.GetCommunityPosts)
			community.POST("/posts", middlewares.AuthMiddleware(), controllers.CreatePost)
		}
	}

	return router
}
