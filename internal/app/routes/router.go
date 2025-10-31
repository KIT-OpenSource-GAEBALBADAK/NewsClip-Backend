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

			// [신규] 인증이 필요한 라우트
			// AuthMiddleware()가 먼저 실행되어 토큰을 검증
			auth.POST("/setup-profile", middlewares.AuthMiddleware(), controllers.SetupProfile)
		}
	}

	return router
}