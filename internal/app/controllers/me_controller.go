package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// === 내 북마크 목록 조회 컨트롤러 ===
func GetMyBookmarks(c *gin.Context) {
	// 1. 미들웨어에서 userID 가져오기
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, _ := userIDValue.(uint)

	// 2. 쿼리 파라미터 파싱
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 3. 서비스 호출
	responseDTO, err := services.GetBookmarkedNewsList(userID, page, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "북마크 목록 조회에 실패했습니다.")
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "북마크 목록 조회 성공",
		"data":    responseDTO,
	})
}
