package controllers

import (
	"errors"
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv" // 문자열 -> 숫자 변환

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// === 뉴스 목록 조회 컨트롤러 ===
func GetNewsList(c *gin.Context) {
	// 1. 파라미터 파싱
	category := c.DefaultQuery("category", "전체")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	// 2. [변경점] UserID 무조건 추출
	// AuthMiddleware가 통과시켰으므로 userID는 반드시 존재합니다.
	userID := c.GetUint("userID")
	// (만약 c.GetUint 헬퍼가 없다면 기존 방식대로 c.Get("userID") 후 형변환 사용)

	// 3. 서비스 호출 (항상 userID 전달)
	responseDTO, err := services.GetNewsList(category, page, size, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "뉴스 조회 실패")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "뉴스 목록 조회 성공",
		"data":    responseDTO,
	})
}

// === 뉴스 상세 조회 컨트롤러 ===
func GetNewsDetail(c *gin.Context) {
	newsIDStr := c.Param("newsId")
	newsID64, err := strconv.ParseUint(newsIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 뉴스 ID입니다.")
		return
	}
	newsID := uint(newsID64)

	// [변경점] UserID 무조건 추출
	userID := c.GetUint("userID")

	// 서비스 호출
	responseDTO, err := services.GetNewsDetail(newsID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendError(c, http.StatusNotFound, "해당 뉴스를 찾을 수 없습니다.")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "뉴스 조회에 실패했습니다.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "뉴스 본문 조회 성공",
		"data":    responseDTO,
	})
}

// === 뉴스 상호작용 컨트롤러 ===
func InteractNews(c *gin.Context) {
	// 1. 미들웨어에서 userID 가져오기
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, _ := userIDValue.(uint)

	// 2. URL에서 newsId 가져오기
	newsIDStr := c.Param("newsId")
	newsID64, err := strconv.ParseUint(newsIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 뉴스 ID입니다.")
		return
	}
	newsID := uint(newsID64)

	// 3. Request Body 바인딩
	var req services.InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	// 4. interaction_type 유효성 검사
	if req.InteractionType != "like" && req.InteractionType != "dislike" {
		utils.SendError(c, http.StatusBadRequest, "interaction_type은 'like' 또는 'dislike'여야 합니다.")
		return
	}

	// 5. 서비스 로직 호출
	responseDTO, err := services.InteractWithNews(userID, newsID, req.InteractionType)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "상호작용 처리에 실패했습니다.")
		return
	}

	// 6. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "상호작용이 처리되었습니다.",
		"data":    responseDTO,
	})
}

// === 뉴스 북마크 토글 컨트롤러 ===
func BookmarkNews(c *gin.Context) {
	// 1. 미들웨어에서 userID 가져오기
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, _ := userIDValue.(uint)

	// 2. URL에서 newsId 가져오기
	newsIDStr := c.Param("newsId")
	newsID64, err := strconv.ParseUint(newsIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 뉴스 ID입니다.")
		return
	}
	newsID := uint(newsID64)

	// 3. 서비스 로직 호출
	isBookmarked, err := services.ToggleBookmark(userID, newsID)
	if err != nil {
		// [수정] 서비스가 ErrRecordNotFound를 처리하므로 컨트롤러에서 제거
		// (대신, newsID가 존재하지 않아 발생하는 FK 에러 등을 여기서 처리)
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			utils.SendError(c, http.StatusNotFound, "해당 뉴스를 찾을 수 없습니다.")
			return
		}

		utils.SendError(c, http.StatusInternalServerError, "북마크 처리에 실패했습니다.")
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "북마크가 처리되었습니다.",
		"data": gin.H{
			"is_bookmarked": isBookmarked,
		},
	})
}
