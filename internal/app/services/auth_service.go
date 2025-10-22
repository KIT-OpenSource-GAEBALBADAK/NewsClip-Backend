package services

import (
	"errors"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/internal/app/utils"

	"gorm.io/gorm"
)

// 회원가입 요청 DTO (Data Transfer Object)
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

func RegisterUser(req RegisterRequest) (models.User, error) {
	// 1. username 중복 체크
	_, err := repositories.FindUserByUsername(req.Username)
	if err == nil { // 에러가 없으면 유저가 존재한다는 의미
		return models.User{}, errors.New("이미 사용 중인 아이디입니다.")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) { // 다른 DB 에러일 경우
		return models.User{}, err
	}

	// 2. 비밀번호 해싱
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return models.User{}, err
	}

	// 3. 유저 모델 생성
	newUser := models.User{
		Name:         req.Name,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Nickname:     req.Nickname,
		ProfileImage: "https://newsclip.duckdns.org/v1/images/default_profile.png", // 기본 프로필
	}

	// 4. DB에 유저 생성
	err = repositories.CreateUser(&newUser)
	if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}
