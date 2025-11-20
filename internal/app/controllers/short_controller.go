package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// === 쇼츠 피드 조회 ===
func GetShortsFeed(c *gin.Context) {
	// 1. Query Parameter (size) 파싱
	sizeStr := c.DefaultQuery("size", "10")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 2. 사용자 ID 확인 (비로그인 유저도 볼 수 있으므로 에러 처리 안 함)
	var userID uint = 0
	if userIDValue, exists := c.Get("userID"); exists {
		if id, ok := userIDValue.(uint); ok {
			userID = id
		}
	}

	// 3. 서비스 호출
	shortsFeed, err := services.GetShortsFeed(size, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "쇼츠 피드 조회 실패")
		return
	}

	// 4. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "쇼츠 피드 조회 성공",
		"data":    shortsFeed,
	})
}

// (참고: InteractShort 컨트롤러도 여기에 나중에 추가됩니다)
