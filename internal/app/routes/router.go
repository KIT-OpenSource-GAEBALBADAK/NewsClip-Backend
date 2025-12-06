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

	setTarget := func(targetType string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("targetType", targetType)
			c.Next()
		}
	}

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
			news.GET("/:newsId/comments", setTarget("news"), controllers.GetComments)
			news.POST("/:newsId/comments", middlewares.AuthMiddleware(), setTarget("news"), controllers.CreateComment)

			// ⭐⭐ [신규] 뉴스 추천 API ⭐⭐
			// GET /v1/news/recommend?size=20
			news.GET("/recommendations/popup", middlewares.AuthMiddleware(), controllers.GetRecommendedNews)
		}

		shorts := v1.Group("/shorts")
		{
			shorts.GET("/", middlewares.AuthMiddlewareOptional(), controllers.GetShortsFeed)
			shorts.POST("/:shortId/interact", middlewares.AuthMiddleware(), controllers.InteractShort)
			shorts.GET("/:shortId/comments", setTarget("short"), controllers.GetComments)
			shorts.POST("/:shortId/comments", middlewares.AuthMiddleware(), setTarget("short"), controllers.CreateComment)
		}

		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			me.GET("/", controllers.GetMyProfile)
			me.POST("/avatar", controllers.UpdateProfile)
			me.GET("/bookmarks", controllers.GetMyBookmarks)

			// 7.4 선호 카테고리 조회
			me.GET("/preferences/categories", controllers.GetPreferredCategories)

			// 7.5 선호 카테고리 설정
			me.PUT("/preferences/categories", controllers.SetPreferredCategories)
			me.GET("/posts", controllers.GetMyPosts)
		}

		community := v1.Group("/community")
		{
			community.GET("/posts", controllers.GetCommunityPosts)
			community.POST("/posts", middlewares.AuthMiddleware(), controllers.CreatePost)
			community.GET("/posts/:postId/comments", setTarget("post"), controllers.GetComments)
			community.POST("/posts/:postId/comments", middlewares.AuthMiddleware(), setTarget("post"), controllers.CreateComment)
			// 게시글 상호작용
			community.POST("/posts/:postId/interact", middlewares.AuthMiddleware(), controllers.InteractPost)
			community.DELETE("/posts/:postId", middlewares.AuthMiddleware(), controllers.DeleteMyPost)
		}
	}

	return router
}
