package routes

import (
	"newsclip/backend/internal/app/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// API V1 그룹
	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/social", controllers.SocialLogin)
			// 여기에 /login, /social, /refresh 라우트도 추가될 예정
		}
	}

	return router
}
