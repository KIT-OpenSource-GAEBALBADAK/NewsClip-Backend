package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv" // 문자열 -> 숫자 변환

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// === [신규] 뉴스 목록 조회 컨트롤러 ===
func GetNewsList(c *gin.Context) {
	// 1. 쿼리 파라미터 파싱
	// (카테고리가 없으면 '전체'를 기본값으로 사용)
	category := c.DefaultQuery("category", "전체")

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

	// 2. (선택적) 사용자 ID 가져오기 (로그인 상태일 수 있으므로)
	// (AuthMiddlewareOptional() 같은 미들웨어가 필요하지만,
	//  우선 GetMyProfile 등에서 사용한 'c.Get("userID")'를 사용)
	var userID uint = 0 // 기본값 0 (비로그인)
	if userIDValue, exists := c.Get("userID"); exists {
		if id, ok := userIDValue.(uint); ok {
			userID = id
		}
	}

	// 3. 서비스 호출
	responseDTO, err := services.GetNewsList(category, page, size, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "뉴스 조회에 실패했습니다.")
		return
	}

	// 4. 성공 응답 (API 명세서 형식에 맞게)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "뉴스 목록 조회 성공",
		"data":    responseDTO, // { "news": [...], "totalItems": ..., "totalPages": ... }
	})
}

// === 뉴스 상세 조회 컨트롤러 ===
func GetNewsDetail(c *gin.Context) {
	// 1. URL 파라미터에서 newsId 추출
	newsIDStr := c.Param("newsId")
	newsID64, err := strconv.ParseUint(newsIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 뉴스 ID입니다.")
		return
	}
	newsID := uint(newsID64)

	// 2. (선택적) 사용자 ID 가져오기
	var userID uint = 0
	if userIDValue, exists := c.Get("userID"); exists {
		if id, ok := userIDValue.(uint); ok {
			userID = id
		}
	}

	// 3. 서비스 호출
	responseDTO, err := services.GetNewsDetail(newsID, userID)
	if err != nil {
		// (gorm.ErrRecordNotFound = DB에 해당 ID가 없음)
		if err == gorm.ErrRecordNotFound {
			utils.SendError(c, http.StatusNotFound, "해당 뉴스를 찾을 수 없습니다.")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "뉴스 조회에 실패했습니다.")
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "뉴스 본문 조회 성공",
		"data":    responseDTO,
	})
}
