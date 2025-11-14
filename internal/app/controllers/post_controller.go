package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
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
