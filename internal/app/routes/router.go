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

			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}

		news := v1.Group("/news")
		{
			news.GET("/", controllers.GetNewsList)
			news.GET("/:newsId", controllers.GetNewsDetail)

			news.POST("/:newsId/interact", middlewares.AuthMiddleware(), controllers.InteractNews)
			news.POST("/:newsId/bookmark", middlewares.AuthMiddleware(), controllers.BookmarkNews)
		}

		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			v1.GET("/", middlewares.AuthMiddleware(), controllers.GetMyProfile)
			v1.POST("/avatar", middlewares.AuthMiddleware(), controllers.UpdateProfile)

			me.GET("/bookmarks", controllers.GetMyBookmarks)
		}

		community := v1.Group("/community")
		{
			community.GET("/posts", controllers.GetCommunityPosts)

			// ⭐ 게시글 작성 추가
			community.POST("/posts", middlewares.AuthMiddleware(), controllers.CreatePost)
		}
	}

	return router
}
