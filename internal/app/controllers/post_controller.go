package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCommunityPosts(c *gin.Context) {
	postType := c.DefaultQuery("type", "all")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	response, err := services.GetCommunityPosts(postType, page, size)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "게시글 목록 조회에 실패했습니다.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "게시글 목록 조회 성공",
		"data":    response,
	})
}

func CreatePost(c *gin.Context) {
	userID := c.GetUint("userID")

	title := c.PostForm("title")
	content := c.PostForm("content")
	category := c.PostForm("category")

	if title == "" || content == "" {
		utils.SendError(c, http.StatusBadRequest, "title과 content는 필수입니다.")
		return
	}

	// 이미지 업로드
	form, err := c.MultipartForm()
	var imageURLs []string

	if err == nil && form.File != nil {
		files := form.File["files"]

		uploadDir := "uploads/posts"
		os.MkdirAll(uploadDir, os.ModePerm)

		for _, file := range files {
			filePath := uploadDir + "/" + file.Filename

			if err := c.SaveUploadedFile(file, filePath); err != nil {
				utils.SendError(c, http.StatusInternalServerError, "이미지 저장 실패")
				return
			}

			imageURLs = append(imageURLs,
				"https://newsclip.duckdns.org/v1/"+filePath)
		}
	}

	// 서비스 호출
	post, err := services.CreatePost(userID, title, content, category, imageURLs)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "게시글 작성에 실패했습니다.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "게시글이 작성되었습니다.",
		"data":    post,
	})
}

// === 게시글 상호작용 컨트롤러 ===
func InteractPost(c *gin.Context) {
	// 1. UserID 확인
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, _ := userIDValue.(uint)

	// 2. PostID 파싱
	postIDStr := c.Param("postId")
	postID64, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 게시글 ID입니다.")
		return
	}
	postID := uint(postID64)

	// 3. Body 파싱 (services.InteractionRequest 재사용)
	// (services 패키지에 InteractionRequest가 public으로 정의되어 있어야 함)
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
	responseDTO, err := services.InteractWithPost(userID, postID, req.InteractionType)
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

// === [5.4] 내가 쓴 게시글 삭제 ===
func DeleteMyPost(c *gin.Context) {

	userID := c.GetUint("userID")

	postIDStr := c.Param("postId")
	postID64, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 게시글 ID입니다.")
		return
	}
	postID := uint(postID64)

	err = services.DeleteMyPost(userID, postID)
	if err != nil {
		if err.Error() == "게시글을 찾을 수 없습니다" {
			utils.SendError(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "본인이 작성한 게시글만 삭제할 수 있습니다" {
			utils.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "게시글 삭제 실패")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "게시글이 삭제되었습니다.",
	})
}
