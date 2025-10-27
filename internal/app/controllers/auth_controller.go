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

// === [추가] ===

// Login 컨트롤러
func Login(c *gin.Context) {
	var req services.LoginRequest
	// 요청 바인딩
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	// 서비스 호출
	response, err := services.LoginUser(req)
	if err != nil {
		// 유저가 없거나 비밀번호가 틀린 경우
		if err.Error() == "아이디 또는 비밀번호가 일치하지 않습니다." {
			utils.SendError(c, http.StatusUnauthorized, err.Error())
			return
		}
		// 기타 서버 에러
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// API 명세서에 맞게 성공 응답 (200 OK)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "로그인 성공",
		"data":    response, // { "accessToken": "...", "refreshToken": "..." }
	})
}

// === [추가] ===

func SocialLogin(c *gin.Context) {
	var req services.SocialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	response, err := services.ProcessSocialLogin(req)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "소셜 로그인 성공",
		"data":    response,
	})
}
