package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"

	"github.com/gin-gonic/gin"
)

func GetMyProfile(c *gin.Context) {
	userID := c.GetUint("userID") // AuthMiddleware에서 넣어줌

	data, err := services.GetMyProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "유저 정보를 가져올 수 없습니다.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "프로필 조회 성공",
		"data":    data,
	})
}
