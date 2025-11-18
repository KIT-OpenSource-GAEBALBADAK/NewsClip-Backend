package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"

	"github.com/gin-gonic/gin"
)

func UpdateNotificationSettings(c *gin.Context) {
	userID := c.GetUint("userID")

	var req services.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	if err := services.UpdateNotificationSettings(userID, req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "알림 설정이 저장되었습니다.",
		"data":    nil,
	})
}
