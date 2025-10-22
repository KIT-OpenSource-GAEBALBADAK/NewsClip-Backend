package controllers

import (
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req services.RegisterRequest
	// 요청 바인딩 및 유효성 검사
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	// 서비스 호출
	user, err := services.RegisterUser(req)
	if err != nil {
		// "이미 사용 중인 아이디입니다." 에러 처리
		if err.Error() == "이미 사용 중인 아이디입니다." {
			utils.SendError(c, http.StatusConflict, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "서버 오류가 발생했습니다.")
		return
	}

	// 성공 응답 (201 Created)
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "회원가입이 완료되었습니다.",
		"data":    gin.H{"user": user},
	})
}
