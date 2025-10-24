package services

import (
	"errors"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/internal/app/utils"
	"time"

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

// --- 회원가입 (기존 코드 끝) ---

// === [추가] ===

// 로그인 요청 DTO
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 로그인 응답 DTO
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// 로그인 로직 처리
func LoginUser(req LoginRequest) (LoginResponse, error) {
	var response LoginResponse

	// 1. 유저 찾기
	user, err := repositories.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다.")
		}
		return response, err // 기타 DB 에러
	}

	// 2. 비밀번호 검증
	isValidPassword := utils.CheckPasswordHash(req.Password, user.PasswordHash)
	if !isValidPassword {
		return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다.")
	}

	// 3. Access Token 생성
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다.")
	}

	// 4. Refresh Token 생성 및 DB 저장
	refreshToken, expiresAt, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다.")
	}

	session := models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := repositories.CreateSession(&session); err != nil {
		return response, errors.New("세션 저장에 실패했습니다.")
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return response, nil
}

// === [추가] Refresh Token 재발급 ===

func RefreshTokens(refreshToken string) (LoginResponse, error) {
	var response LoginResponse

	// 1. Refresh Token 으로 세션 조회
	session, err := repositories.FindSessionByToken(refreshToken)
	if err != nil {
		return response, errors.New("유효하지 않은 Refresh Token 입니다.")
	}

	// 2. 만료 검증
	if time.Now().After(session.ExpiresAt) {
		return response, errors.New("Refresh Token 이 만료되었습니다. 다시 로그인 해주세요.")
	}

	// 3. 유저 정보 조회
	user, err := repositories.FindUserByID(session.UserID)
	if err != nil {
		return response, errors.New("유저 정보를 찾을 수 없습니다.")
	}

	// 4. Access Token 재발급
	newAccess, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return response, errors.New("Access Token 생성 실패")
	}

	// 5. Refresh Token 재발급 (Token Rotation)
	newRefresh, newExp, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, errors.New("Refresh Token 생성 실패")
	}

	// 6. 세션 DB 갱신
	session.RefreshToken = newRefresh
	session.ExpiresAt = newExp
	if err := repositories.UpdateSession(&session); err != nil {
		return response, errors.New("세션 갱신 실패")
	}

	// 7. 반환 DTO 구성
	response.AccessToken = newAccess
	response.RefreshToken = newRefresh
	return response, nil
}
