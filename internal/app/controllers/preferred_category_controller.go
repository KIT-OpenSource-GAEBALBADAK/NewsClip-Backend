package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"

	"github.com/gin-gonic/gin"
)

func GetPreferredCategories(c *gin.Context) {
	userID := c.GetUint("userID")

	categories, err := services.GetPreferredCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "선호 카테고리를 가져올 수 없습니다.",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "선호 카테고리 조회 성공",
		"data": gin.H{
			"categories": categories,
		},
	})
}
