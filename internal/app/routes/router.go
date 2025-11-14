package routes

import (
	"newsclip/backend/internal/app/controllers"
	"newsclip/backend/internal/app/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 업로드된 파일에 접근할 수 있도록 정적 라우트 설정
	// /v1/uploads/profiles/1_image.png 로 요청하면
	// ./uploads/profiles/1_image.png 파일을 서빙
	router.Static("/v1/uploads", "./uploads")

	// (추가) 정적 자산(기본 이미지 등) 서빙
	// URL /v1/images/ 로 오는 요청을 ./static/images/ 디렉토리와 연결
	router.Static("/v1/images", "./static/images")

	// API V1 그룹
	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			// 인증이 필요 없는 라우트
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/social", controllers.SocialLogin)
			auth.POST("/refresh", controllers.RefreshToken)
			auth.POST("/check-username", controllers.CheckUsername)

			// 인증이 필요한 라우트
			// AuthMiddleware()가 먼저 실행되어 토큰을 검증
			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}

		news := v1.Group("/news")
		{
			// (참고: AuthMiddlewareOptional() 같은 것이 필요)
			// 우선 인증 없이 라우트 등록
			news.GET("/", controllers.GetNewsList)
			// === /:newsId 엔드포인트 연결 ===
			// (참고: 인증 없이도 볼 수 있어야 하므로 미들웨어 제외)
			news.GET("/:newsId", controllers.GetNewsDetail)
			// (인증 필요) P3. 뉴스 상호작용 (좋아요/싫어요)
			news.POST("/:newsId/interact", middlewares.AuthMiddleware(), controllers.InteractNews)
			// 뉴스 북마크 - 인증 필요
			news.POST("/:newsId/bookmark", middlewares.AuthMiddleware(), controllers.BookmarkNews)
		}

		me := v1.Group("/me", middlewares.AuthMiddleware())
		{
			// 내 프로필 조회 추가
			v1.GET("/", middlewares.AuthMiddleware(), controllers.GetMyProfile)
			// 내 프로필 수정
			v1.POST("/avatar", middlewares.AuthMiddleware(), controllers.UpdateProfile)
			// === 내 북마크 조회 ===
			me.GET("/bookmarks", controllers.GetMyBookmarks)

			// ... (기타 /preferences/categories 등) ...
		}

	}

	return router
}
