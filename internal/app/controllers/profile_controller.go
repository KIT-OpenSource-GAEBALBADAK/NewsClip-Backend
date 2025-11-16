package controllers

import (
	"fmt"
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func GetMyProfile(c *gin.Context) {
	userID := c.GetUint("userID") // AuthMiddleware에서 넣어줌

	data, err := services.GetMyProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "유저 정보를 가져올 수 없습니다.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "프로필 조회 성공",
		"data":    data,
	})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	nickname := c.PostForm("nickname")
	file, _ := c.FormFile("file")

	// 이미지 처리
	var imageURL *string
	if file != nil {
		uploadDir := "uploads/profiles"
		os.MkdirAll(uploadDir, os.ModePerm)
		filePath := fmt.Sprintf("%s/%d_%s", uploadDir, userID, file.Filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.SendError(c, http.StatusInternalServerError, "파일 저장에 실패했습니다.")
			return
		}

		url := "https://newsclip.duckdns.org/v1/" + filePath
		imageURL = &url
	}

	user, err := services.UpdateProfile(userID, nickname, imageURL)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

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
