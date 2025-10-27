package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// === [추가] ===

// 소셜 로그인 요청 DTO
type SocialLoginRequest struct {
	Provider string `json:"provider" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// 카카오 유저 정보 응답 DTO
type KakaoUserResponse struct {
	ID           int64 `json:"id"`
	KakaoAccount struct {
		Profile struct {
			Nickname        string `json:"nickname"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"profile"`
	} `json:"kakao_account"`
}

// 카카오 서버에 유저 정보 요청
func getKakaoUserInfo(token string) (*KakaoUserResponse, error) {
	req, err := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("카카오 서버 에러: %s", string(body))
	}

	var kakaoUser KakaoUserResponse
	if err := json.Unmarshal(body, &kakaoUser); err != nil {
		return nil, err
	}

	return &kakaoUser, nil
}

// 소셜 로그인 전체 로직
func ProcessSocialLogin(req SocialLoginRequest) (LoginResponse, error) {
	var response LoginResponse
	var user models.User
	var err error

	if req.Provider == "kakao" {
		kakaoUser, err := getKakaoUserInfo(req.Token)
		if err != nil {
			return response, err
		}

		providerID := fmt.Sprintf("%d", kakaoUser.ID)

		// 1. 이미 가입된 유저인지 확인
		user, err = repositories.FindUserBySocial(req.Provider, providerID)

		// 2. 가입되지 않은 유저라면, 새로 생성 (회원가입)
		if errors.Is(err, gorm.ErrRecordNotFound) {

			provider := "kakao"
			pID := providerID

			newUser := models.User{
				Name:         kakaoUser.KakaoAccount.Profile.Nickname, // 실명 대신 닉네임 사용
				Nickname:     kakaoUser.KakaoAccount.Profile.Nickname,
				ProfileImage: kakaoUser.KakaoAccount.Profile.ProfileImageURL,
				Provider:     &provider,
				ProviderID:   &pID,
			}

			if err := repositories.CreateUser(&newUser); err != nil {
				return response, err
			}
			user = newUser
		} else if err != nil {
			return response, err // 기타 DB 에러
		}

	} else {
		return response, errors.New("지원하지 않는 소셜 로그인입니다.")
	}

	// 3. 우리 서비스의 JWT 토큰 발급
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return response, err
	}

	refreshToken, expiresAt, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, err
	}

	session := models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := repositories.CreateSession(&session); err != nil {
		return response, err
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return response, nil
}
