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

	// 2. [신규] Query Parameter (cursorId) 파싱
	// 값이 없으면 기본값 0 -> 첫 페이지(최신) 조회
	cursorStr := c.Query("cursorId")
	var cursorID uint = 0

	if cursorStr != "" {
		parsedID, err := strconv.ParseUint(cursorStr, 10, 32)
		if err == nil {
			cursorID = uint(parsedID)
		}
	}

	// 3. 사용자 ID 확인 (Optional)
	var userID uint = 0
	if userIDValue, exists := c.Get("userID"); exists {
		if id, ok := userIDValue.(uint); ok {
			userID = id
		}
	}

	// 4. 서비스 호출 (cursorID 전달)
	shortsFeed, err := services.GetShortsFeed(size, cursorID, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "쇼츠 피드 조회 실패")
		return
	}

	// 5. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "쇼츠 피드 조회 성공",
		"data":    shortsFeed,
	})
}

// === 쇼츠 상호작용 컨트롤러 ===
func InteractShort(c *gin.Context) {
	// 1. UserID 확인 (AuthMiddleware 필수)
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, _ := userIDValue.(uint)

	// 2. ShortID 파싱
	shortIDStr := c.Param("shortId")
	shortID64, err := strconv.ParseUint(shortIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 쇼츠 ID입니다.")
		return
	}
	shortID := uint(shortID64)

	// 3. Body 파싱 (news_service에 있는 구조체 재사용)
	var req services.InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	if req.InteractionType != "like" && req.InteractionType != "dislike" {
		utils.SendError(c, http.StatusBadRequest, "interaction_type은 'like' 또는 'dislike'여야 합니다.")
		return
	}

	// 4. 서비스 호출
	responseDTO, err := services.InteractWithShort(userID, shortID, req.InteractionType)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "상호작용 처리에 실패했습니다.")
		return
	}

	// 5. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "상호작용이 처리되었습니다.",
		"data":    responseDTO,
	})
}
