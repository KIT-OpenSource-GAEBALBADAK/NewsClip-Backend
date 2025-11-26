package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"

	"github.com/gin-gonic/gin"
)

type SetPreferredCategoriesRequest struct {
	Categories []string `json:"categories" binding:"required"`
}

// 7.4 조회
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

// ⭐⭐ 7.5 선호 카테고리 설정 ⭐⭐
func SetPreferredCategories(c *gin.Context) {
	var req SetPreferredCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "categories 필드가 필요합니다.",
		})
		return
	}

	userID := c.GetUint("userID")

	err := services.SetPreferredCategories(userID, req.Categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "선호 카테고리 저장 실패",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "선호 카테고리가 저장되었습니다.",
		"data":    nil,
	})
}
