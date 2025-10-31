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
	"time"

	"gorm.io/gorm"
)

// === [수정] 회원가입 요청 DTO === 25/10/31
type RegisterRequest struct {
	// Name     string `json:"name" binding:"required"` // 삭제
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	// Nickname string `json:"nickname" binding:"required"` // 삭제
}

// === [수정] 회원가입 로직 === 25/10/31
func RegisterUser(req RegisterRequest) (models.User, error) {
	// 1. username 중복 체크
	_, err := repositories.FindUserByUsername(req.Username)
	if err == nil {
		return models.User{}, errors.New("이미 사용 중인 아이디입니다.")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}

	// 2. 비밀번호 해싱
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return models.User{}, err
	}

	// 3. 유저 모델 생성 (Nickname, ProfileImage는 nil)
	// model에서 포인터 타입이므로, 변수에 담아서 주소를 넘겨줌
	username := req.Username
	pwhash := hashedPassword

	newUser := models.User{
		Username:     &username,
		PasswordHash: &pwhash,
		// Nickname, ProfileImage는 기본값(NULL)
		// Name 필드 삭제됨
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

// === [수정] 로그인 로직 === 25/10/31
func LoginUser(req LoginRequest) (LoginResponse, error) {
	var response LoginResponse

	// 1. 유저 찾기
	user, err := repositories.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다.")
		}
		return response, err
	}

	// 2. 비밀번호 검증 (소셜 로그인 유저는 PasswordHash가 nil일 수 있음)
	if user.PasswordHash == nil {
		return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다.")
	}
	isValidPassword := utils.CheckPasswordHash(req.Password, *user.PasswordHash)
	if !isValidPassword {
		return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다.")
	}

	// 3. Access Token 생성 (Nickname이 nil일 수 있음)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	accessToken, err := utils.GenerateAccessToken(user.ID, nickname)
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

// === [수정] 소셜 로그인 전체 로직 ===
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

		// 2. 가입되지 않은 유저라면, 새로 생성 (Nickname, ProfileImage는 NULL)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			provider := "kakao"
			pID := providerID

			newUser := models.User{
				Provider:   &provider,
				ProviderID: &pID,
				// Name, Nickname, ProfileImage 모두 NULL
			}

			if err := repositories.CreateUser(&newUser); err != nil {
				return response, err
			}
			user = newUser
		} else if err != nil {
			return response, err
		}

	} else {
		return response, errors.New("지원하지 않는 소셜 로그인입니다.")
	}

	// 3. 우리 서비스의 JWT 토큰 발급 (Nickname이 nil일 수 있음)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	accessToken, err := utils.GenerateAccessToken(user.ID, nickname)
	if err != nil {
		return response, err
	}

	// 4. Refresh Token 생성 및 세션 저장
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

// === [수정] Refresh Token 재발급 ===
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

	// 4. Access Token 재발급 (Nickname nil 처리)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	newAccess, err := utils.GenerateAccessToken(user.ID, nickname)
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

// === [신규] 아이디 중복 체크 ===
type CheckUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

func CheckUsernameAvailability(req CheckUsernameRequest) (bool, error) {
	_, err := repositories.FindUserByUsername(req.Username)
	if err == nil { // 유저가 존재함
		return false, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) { // 유저가 존재하지 않음
		return true, nil
	}
	return false, err // 기타 DB 에러
}

// === [신규] 최초 프로필 설정 ===
// (컨트롤러에서 파일 처리 후 imageURL을 받아옴)
func SetupProfile(userID uint, nickname string, imageURL string) (models.User, error) {
	// 1. 유저 정보 가져오기
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return user, errors.New("유저를 찾을 수 없습니다.")
	}

	// 2. 이미 설정했는지 확인
	if user.Nickname != nil {
		return user, errors.New("이미 프로필이 설정되었습니다.")
	}

	// 3. DB 업데이트
	err = repositories.UpdateUserProfile(&user, nickname, imageURL)
	if err != nil {
		return user, errors.New("프로필 업데이트에 실패했습니다.")
	}

	user.Nickname = &nickname
	user.ProfileImage = &imageURL
	return user, nil
}