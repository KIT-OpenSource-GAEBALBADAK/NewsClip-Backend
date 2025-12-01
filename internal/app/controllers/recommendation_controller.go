package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 추천 뉴스 조회 컨트롤러
// GET /v1/news/recommend?size=5
func GetRecommendedNews(c *gin.Context) {
	// 1. 인증된 사용자 ID 가져오기
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 잘못되었습니다.")
		return
	}

	// 2. 추천 개수(size) 파라미터 파싱 (기본값: 5)
	sizeStr := c.DefaultQuery("size", "5")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 5
	}
	if size > 50 { // 안전한 상한값
		size = 50
	}

	// 3. 서비스 호출
	resp, err := services.GetRecommendedNews(userID, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "추천 뉴스 조회에 실패했습니다.")
		return
	}

	// 4. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "추천 뉴스 조회 성공",
		"data":    resp,
	})
}
