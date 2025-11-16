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
