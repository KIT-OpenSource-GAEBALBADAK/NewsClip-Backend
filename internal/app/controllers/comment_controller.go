package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 댓글 작성 요청 Body
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// [헬퍼 함수] 타겟 타입에 따라 올바른 ID 파라미터를 추출
func getTargetID(c *gin.Context) (uint, error) {
	targetType := c.GetString("targetType")
	var idStr string

	// 타입별로 파라미터 이름 매핑
	switch targetType {
	case "news":
		idStr = c.Param("newsId") // router에서 :newsId 라고 했으므로
	case "short":
		idStr = c.Param("shortId") // router에서 :shortId 라고 했으므로
	case "post":
		idStr = c.Param("postId") // router에서 :postId 라고 했으므로
	default:
		idStr = c.Param("id") // 기본값
	}

	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}

// === 댓글 작성 컨트롤러 ===
func CreateComment(c *gin.Context) {
	targetType := c.GetString("targetType")

	// 헬퍼 함수 사용
	targetID, err := getTargetID(c)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 ID입니다.")
		return
	}

	// 2. UserID 확인
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "로그인이 필요합니다.")
		return
	}
	userID := userIDValue.(uint)

	// 3. Body 파싱
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "내용을 입력해주세요.")
		return
	}

	// 4. 서비스 호출
	commentID, err := services.CreateComment(targetType, targetID, userID, req.Content)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "댓글 작성 실패")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "댓글이 등록되었습니다.",
		"data":    gin.H{"comment_id": commentID},
	})
}

// === 댓글 목록 조회 컨트롤러 ===
func GetComments(c *gin.Context) {
	targetType := c.GetString("targetType")

	// 헬퍼 함수 사용
	targetID, err := getTargetID(c)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 ID입니다.")
		return
	}

	comments, err := services.GetComments(targetType, targetID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "댓글 조회 실패")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "댓글 목록 조회 성공",
		"data":    comments,
	})
}
