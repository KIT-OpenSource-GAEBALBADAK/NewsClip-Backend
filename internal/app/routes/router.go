package routes

import (
	"newsclip/backend/internal/app/controllers"
	"newsclip/backend/internal/app/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 업로드된 파일에 접근할 수 있도록 정적 라우트 설정
	router.Static("/v1/uploads", "./uploads")

	// 정적 자산(기본 이미지 등)
	router.Static("/v1/images", "./static/images")

	// API V1 그룹
	v1 := router.Group("/v1")
	{
		/* ===========================
		         AUTH
		=========================== */
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/social", controllers.SocialLogin)
			auth.POST("/refresh", controllers.RefreshToken)
			auth.POST("/check-username", controllers.CheckUsername)

			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}

		/* ===========================
		         NEWS
		=========================== */
		news := v1.Group("/news")
		{
			news.GET("/", controllers.GetNewsList)
			news.GET("/:newsId", controllers.GetNewsDetail)

			news.POST("/:newsId/interact", middlewares.AuthMiddleware(), controllers.InteractNews)
			news.POST("/:newsId/bookmark", middlewares.AuthMiddleware(), controllers.BookmarkNews)
		}

		/* ===========================
		          ME
		=========================== */
		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			v1.GET("/", middlewares.AuthMiddleware(), controllers.GetMyProfile)
			v1.POST("/avatar", middlewares.AuthMiddleware(), controllers.UpdateProfile)

			me.GET("/bookmarks", controllers.GetMyBookmarks)
		}

		/* ===========================
		       COMMUNITY (NEW)
		=========================== */
		community := v1.Group("/community")
		{
			// ⭐ 게시글 목록 조회 추가됨
			community.GET("/posts", controllers.GetCommunityPosts)
		}
	}

	return router
}
