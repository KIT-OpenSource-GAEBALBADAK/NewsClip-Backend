package controllers

import (
	"fmt"
	"net/http"
	"newsclip/backend/internal/app/services"
	"newsclip/backend/internal/app/utils"
	"os"

	"github.com/gin-gonic/gin"
)

// === 회원가입 ===
func Register(c *gin.Context) {
	var req services.RegisterRequest // 수정된 DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	user, err := services.RegisterUser(req)
	if err != nil {
		if err.Error() == "이미 사용 중인 아이디입니다." {
			utils.SendError(c, http.StatusConflict, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "서버 오류가 발생했습니다.")
		return
	}

	// === [수정] 성공 응답 (API 명세서) ===
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "회원가입이 완료되었습니다.",
		"data": gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username, // 포인터라도 gin.H가 처리해줌
				"role":     user.Role,
			},
		},
	})
}

// === 로그인 ===
func Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	response, err := services.LoginUser(req)
	if err != nil {
		if err.Error() == "아이디 또는 비밀번호가 일치하지 않습니다." {
			utils.SendError(c, http.StatusUnauthorized, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "로그인 성공",
		"data":    response,
	})
}

// === 소셜 로그인 ===
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

// === 토큰 재발급 ===
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Refresh Token 이 필요합니다.")
		return
	}

	response, err := services.RefreshTokens(req.RefreshToken)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "토큰 재발급 성공",
		"data":    response,
	})
}

// === 아이디 중복 체크 ===
func CheckUsername(c *gin.Context) {
	var req services.CheckUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "잘못된 요청 형식입니다.")
		return
	}

	isAvailable, err := services.CheckUsernameAvailability(req)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "서버 오류가 발생했습니다.")
		return
	}

	if isAvailable {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "사용 가능한 아이디입니다.",
			"data":    gin.H{"isAvailable": true},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "이미 사용 중인 아이디입니다.",
			"data":    gin.H{"isAvailable": false},
		})
	}
}

// === 최초 프로필 설정 ===
func SetupProfile(c *gin.Context) {
	// 1. 미들웨어에서 userID 가져오기
	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 없습니다.")
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "인증 정보가 잘못되었습니다.")
		return
	}

	// 2. Form 데이터 파싱
	nickname := c.PostForm("nickname")
	if nickname == "" {
		utils.SendError(c, http.StatusBadRequest, "닉네임은 필수입니다.")
		return
	}

	file, err := c.FormFile("file") // "file"은 폼 필드 이름

	// 3. 이미지 처리
	var imageURL string
	defaultImageURL := "https://newsclip.duckdns.org/v1/images/default_profile.png"

	if err != nil { // 파일이 없는 경우 (http.ErrMissingFile 등)
		imageURL = defaultImageURL
	} else {
		// (간단한 로컬 저장 예시. 프로덕션에서는 S3/GCS 사용)
		uploadDir := "uploads/profiles"
		// ./uploads/profiles 경로 생성
		os.MkdirAll(uploadDir, os.ModePerm)

		// 파일 경로: uploads/profiles/유저ID_파일명
		filePath := fmt.Sprintf("%s/%d_%s", uploadDir, userID, file.Filename)

		// 파일 저장
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.SendError(c, http.StatusInternalServerError, "파일 저장에 실패했습니다.")
			return
		}
		// 클라이언트가 접근할 수 있는 URL (router.go에서 설정한 /v1/uploads/...)
		imageURL = "https://newsclip.duckdns.org/v1/" + filePath
	}

	// 4. 서비스 호출
	user, err := services.SetupProfile(userID, nickname, imageURL)
	if err != nil {
		if err.Error() == "이미 프로필이 설정되었습니다." {
			utils.SendError(c, http.StatusConflict, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "프로필 설정이 완료되었습니다.",
		"data": gin.H{
			"user": gin.H{
				"id":            user.ID,
				"nickname":      user.Nickname,
				"profile_image": user.ProfileImage,
			},
		},
	})
}

// 요청 바디 구조체
type EmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required"` // "signup" or "reset"
}

// 인증번호 전송 컨트롤러
func SendEmailCode(c *gin.Context) {
	var req EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "이메일 형식이 올바르지 않습니다.")
		return
	}

	err := services.SendEmailVerification(req.Email, req.Type)
	if err != nil {
		if err.Error() == "already_exists" {
			utils.SendError(c, http.StatusConflict, "이미 가입된 이메일입니다.")
			return
		}
		if err.Error() == "not_found" {
			utils.SendError(c, http.StatusNotFound, "가입되지 않은 이메일입니다.")
			return
		}
		if err.Error() == "invalid_type" {
			utils.SendError(c, http.StatusBadRequest, "잘못된 인증 타입입니다.")
			return
		}

		// 그 외(Redis, SMTP 에러 등)
		utils.SendError(c, http.StatusInternalServerError, "인증번호 전송에 실패했습니다: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "인증코드가 전송되었습니다. 이메일을 확인해주세요.",
		"data": gin.H{
			"expiration_time": 180,
		},
	})
}
