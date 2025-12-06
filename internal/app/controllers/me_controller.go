package controllers

import (
	"fmt"
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==========================
// 1. 내 프로필 조회 (통계 포함)
// ==========================
func GetMyProfile(c *gin.Context) {
	// 1. UserID 추출 (AuthMiddleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID := userIDValue.(uint)

	// 2. 서비스 호출
	data, err := services.GetMyProfile(userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "유저 정보를 가져올 수 없습니다.")
		return
	}

	// 3. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "프로필 조회 성공",
		"data":    data,
	})
}

// ==========================
// 2. 내 프로필 수정 (닉네임, 이미지)
// ==========================
func UpdateProfile(c *gin.Context) {
	// 1. UserID 추출
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID := userIDValue.(uint)

	nickname := c.PostForm("nickname")
	file, _ := c.FormFile("file")

	// 2. 이미지 파일 처리
	var imageURL *string
	if file != nil {
		uploadDir := "uploads/profiles"
		os.MkdirAll(uploadDir, os.ModePerm)

		// 파일명 충돌 방지를 위해 userID_파일명 형식 사용
		filePath := fmt.Sprintf("%s/%d_%s", uploadDir, userID, file.Filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.SendError(c, http.StatusInternalServerError, "파일 저장에 실패했습니다.")
			return
		}

		// 도메인 주소 포함
		url := "https://newsclip.duckdns.org/v1/" + filePath
		imageURL = &url
	}

	// 3. 서비스 호출
	user, err := services.UpdateProfile(userID, nickname, imageURL)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 4. 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "내 정보가 변경되었습니다.",
		"data": gin.H{
			"id":            user.ID,
			"nickname":      user.Nickname,
			"profile_image": user.ProfileImage,
			"updated_at":    user.UpdatedAt,
		},
	})
}

// ==========================
// 3. 내 북마크 목록 조회
// ==========================
func GetMyBookmarks(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID := userIDValue.(uint)

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

	responseDTO, err := services.GetBookmarkedNewsList(userID, page, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "북마크 목록 조회에 실패했습니다.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "북마크 목록 조회 성공",
		"data":    responseDTO,
	})
}

// ==========================
// 7.7 내가 쓴 게시글 목록 조회
// ==========================
func GetMyPosts(c *gin.Context) {

	userID := c.GetUint("userID")

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

	resp, err := services.GetMyPosts(userID, page, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "내 게시글 목록 조회 실패")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "내가 쓴 게시글 목록 조회 성공",
		"data":    resp,
	})
}

// ==========================
// 7.8 내가 쓴 댓글 목록 조회
// ==========================
func GetMyComments(c *gin.Context) {

	userID := c.GetUint("userID")

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

	resp, err := services.GetMyComments(userID, page, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "내 댓글 목록 조회 실패")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "내가 쓴 댓글 목록 조회 성공",
		"data":    resp,
	})
}
