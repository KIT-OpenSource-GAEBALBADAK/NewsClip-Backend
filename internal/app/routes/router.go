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

	// 헬퍼: Target Type을 설정해주는 미들웨어
	setTarget := func(targetType string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("targetType", targetType) // 컨트롤러에서 c.GetString("targetType")으로 사용
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

			// 그룹에 미들웨어 적용
			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}

		news := v1.Group("/news")
		{
			news.GET("/", controllers.GetNewsList)
			news.GET("/:newsId", controllers.GetNewsDetail)

			news.POST("/:newsId/interact", middlewares.AuthMiddleware(), controllers.InteractNews)
			news.POST("/:newsId/bookmark", middlewares.AuthMiddleware(), controllers.BookmarkNews)

			// :id -> :newsId 로 변경 (이름 통일)
			news.GET("/:newsId/comments", setTarget("news"), controllers.GetComments)
			news.POST("/:newsId/comments", middlewares.AuthMiddleware(), setTarget("news"), controllers.CreateComment)
		}

		shorts := v1.Group("/shorts")
		{
			// 1. 피드 조회 (비로그인 가능 - Optional 미들웨어)
			shorts.GET("/", middlewares.AuthMiddlewareOptional(), controllers.GetShortsFeed)
			// 2. 상호작용 (로그인 필수 - AuthMiddleware)
			shorts.POST("/:shortId/interact", middlewares.AuthMiddleware(), controllers.InteractShort)

			// :id -> :shortId 로 변경
			shorts.GET("/:shortId/comments", setTarget("short"), controllers.GetComments)
			shorts.POST("/:shortId/comments", middlewares.AuthMiddleware(), setTarget("short"), controllers.CreateComment)
		}

		// me 그룹에 인증 미들웨어를 붙임
		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			me.GET("/", controllers.GetMyProfile)
			me.POST("/avatar", controllers.UpdateProfile)
			me.GET("/bookmarks", controllers.GetMyBookmarks)
		}

		community := v1.Group("/community")
		{
			community.GET("/posts", controllers.GetCommunityPosts)
			community.POST("/posts", middlewares.AuthMiddleware(), controllers.CreatePost)

			// 게시글 댓글 (주의: 여기 id는 post_id)
			// 명확하게 :postId 사용
			community.GET("/posts/:postId/comments", setTarget("post"), controllers.GetComments)
			community.POST("/posts/:postId/comments", middlewares.AuthMiddleware(), setTarget("post"), controllers.CreateComment)
		}
	}

	return router
}
