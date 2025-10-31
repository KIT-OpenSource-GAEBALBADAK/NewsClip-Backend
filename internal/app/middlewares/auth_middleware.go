package middlewares

import (
	"net/http"
	"newsclip/backend/internal/app/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware는 JWT 토큰을 검증하고 컨텍스트에 userID를 설정합니다.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "인증 헤더가 필요합니다.")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.SendError(c, http.StatusUnauthorized, "인증 헤더 형식이 잘못되었습니다.")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString) // 이 함수는 utils/jwt.go에 만들어야 합니다.
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "유효하지 않은 토큰입니다.")
			c.Abort()
			return
		}

		// 컨텍스트에 유저 정보 설정
		c.Set("userID", claims.UserID)
		c.Set("userNickname", claims.Nickname)

		c.Next()
	}
}