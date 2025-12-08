package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/internal/app/utils"
	"newsclip/backend/pkg/email"
	"newsclip/backend/pkg/redis"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// === 회원가입 요청 DTO === 25/10/31
type RegisterRequest struct {
	// Name     string `json:"name" binding:"required"` // 삭제
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	// Nickname string `json:"nickname" binding:"required"` // 삭제
}

// === 회원가입 로직 === 25/10/31
func RegisterUser(req RegisterRequest) (models.User, error) {
	// 1. username 중복 체크
	_, err := repositories.FindUserByUsername(req.Username)
	if err == nil {
		return models.User{}, errors.New("이미 사용 중인 아이디입니다")
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

// === 로그인 로직 === 25/10/31
func LoginUser(req LoginRequest) (LoginResponse, error) {
	var response LoginResponse

	// 1. 유저 찾기
	user, err := repositories.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다")
		}
		return response, err
	}

	// 2. 비밀번호 검증 (소셜 로그인 유저는 PasswordHash가 nil일 수 있음)
	if user.PasswordHash == nil {
		return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다")
	}
	isValidPassword := utils.CheckPasswordHash(req.Password, *user.PasswordHash)
	if !isValidPassword {
		return response, errors.New("아이디 또는 비밀번호가 일치하지 않습니다")
	}

	// 3. Access Token 생성 (Nickname이 nil일 수 있음)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	accessToken, err := utils.GenerateAccessToken(user.ID, nickname)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다")
	}

	// 4. Refresh Token 생성 및 DB 저장
	refreshToken, expiresAt, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다")
	}

	session := models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := repositories.CreateSession(&session); err != nil {
		return response, errors.New("세션 저장에 실패했습니다")
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return response, nil
}

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

// (구글) 25/11/04
type GoogleUserResponse struct {
	Sub      string `json:"sub"`     // 구글 유저 고유 ID
	Audience string `json:"aud"`     // 토큰 발급 대상 (우리 앱 Client ID)
	Email    string `json:"email"`   // (참고용)
	Name     string `json:"name"`    // (참고용)
	Picture  string `json:"picture"` // (참고용)
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

// === (구글) 유저 정보 요청 === 25/11/04
// 구글은 ID Token을 검증하는 'tokeninfo' 엔드포인트를 사용합니다.
func getGoogleUserInfo(token string) (*GoogleUserResponse, error) {
	// 1. 구글 tokeninfo 엔드포인트에 GET 요청
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + token)
	if err != nil {
		return nil, fmt.Errorf("구글 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("구글 응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("구글 토큰 검증 에러: %s", string(body))
	}

	// 2. 응답 파싱
	var googleUser GoogleUserResponse
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("구글 응답 파싱 실패: %w", err)
	}

	// 3. [보안] Audience 검증 (이 토큰이 우리 앱에 발급된게 맞는지)
	googleClientID := config.GetEnv("GOOGLE_CLIENT_ID")
	if googleUser.Audience != googleClientID {
		return nil, errors.New("유효하지 않은 구글 토큰입니다. (Audience 불일치)")
	}

	return &googleUser, nil
}

// === 소셜 로그인 전체 로직 함수 ===

func ProcessSocialLogin(req SocialLoginRequest) (LoginResponse, error) {
	var response LoginResponse
	var user models.User
	var err error
	var providerID string // 카카오, 구글 ID를 담을 변수

	// 1. Provider에 따라 분기
	if req.Provider == "kakao" {
		kakaoUser, err := getKakaoUserInfo(req.Token)
		if err != nil {
			return response, err
		}
		providerID = fmt.Sprintf("%d", kakaoUser.ID) // 카카오 ID

	} else if req.Provider == "google" {
		googleUser, err := getGoogleUserInfo(req.Token)
		if err != nil {
			return response, err
		}
		providerID = googleUser.Sub // 구글 고유 ID는 'sub' 필드

	} else {
		return response, errors.New("지원하지 않는 소셜 로그인입니다")
	}

	// --- [이하 로직은 공통] ---

	// 2. 이미 가입된 유저인지 확인
	user, err = repositories.FindUserBySocial(req.Provider, providerID)

	// 3. 가입되지 않은 유저라면, 새로 생성 (Nickname, ProfileImage는 NULL)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		provider := req.Provider // "kakao" 또는 "google"
		pID := providerID

		newUser := models.User{
			Provider:   &provider,
			ProviderID: &pID,
			// Name, Nickname, ProfileImage 모두 NULL (최초 프로필 설정 필요)
		}

		if err := repositories.CreateUser(&newUser); err != nil {
			return response, err
		}
		user = newUser
	} else if err != nil {
		return response, err // 기타 DB 에러
	}

	// 4. 우리 서비스의 JWT 토큰 발급 (Nickname이 nil일 수 있음)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	accessToken, err := utils.GenerateAccessToken(user.ID, nickname)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다")
	}

	// 5. Refresh Token 생성 및 세션 저장
	refreshToken, expiresAt, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, errors.New("토큰 생성에 실패했습니다")
	}
	session := models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := repositories.CreateSession(&session); err != nil {
		return response, errors.New("세션 저장에 실패했습니다")
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return response, nil
}

// === Refresh Token 재발급 ===
func RefreshTokens(refreshToken string) (LoginResponse, error) {
	var response LoginResponse

	// 1. Refresh Token 으로 세션 조회
	session, err := repositories.FindSessionByToken(refreshToken)
	if err != nil {
		return response, errors.New("유효하지 않은 Refresh Token 입니다")
	}

	// 2. 만료 검증
	if time.Now().After(session.ExpiresAt) {
		return response, errors.New("refresh token 이 만료되었습니다. 다시 로그인 해주세요")
	}

	// 3. 유저 정보 조회
	user, err := repositories.FindUserByID(session.UserID)
	if err != nil {
		return response, errors.New("유저 정보를 찾을 수 없습니다")
	}

	// 4. Access Token 재발급 (Nickname nil 처리)
	var nickname string
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	newAccess, err := utils.GenerateAccessToken(user.ID, nickname)
	if err != nil {
		return response, errors.New("access token 생성 실패")
	}

	// 5. Refresh Token 재발급 (Token Rotation)
	newRefresh, newExp, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response, errors.New("refresh token 생성 실패")
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

// === 아이디 중복 체크 ===
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

// === 최초 프로필 설정 ===
// (컨트롤러에서 파일 처리 후 imageURL을 받아옴)
func SetupProfile(userID uint, nickname string, imageURL string) (models.User, error) {
	// 1. 유저 정보 가져오기
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return user, errors.New("유저를 찾을 수 없습니다")
	}

	// 2. 이미 설정했는지 확인
	if user.Nickname != nil {
		return user, errors.New("이미 프로필이 설정되었습니다")
	}

	// 3. DB 업데이트
	err = repositories.UpdateUserProfile(&user, nickname, imageURL)
	if err != nil {
		return user, errors.New("프로필 업데이트에 실패했습니다")
	}

	user.Nickname = &nickname
	user.ProfileImage = &imageURL
	return user, nil
}

// === 인증번호 전송 서비스 ===
func SendEmailVerification(emailAddr string, authType string) error {
	// 1. 유저 존재 여부 확인
	_, err := repositories.FindUserByUsername(emailAddr) // Username == Email 이라고 가정

	if authType == "signup" {
		// 회원가입: 이미 유저가 있으면 에러 (409)
		if err == nil {
			return errors.New("already_exists") // 컨트롤러에서 409 처리
		}
	} else if authType == "reset" {
		// 비번찾기: 유저가 없으면 에러 (404)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("not_found") // 컨트롤러에서 404 처리
		}
	} else {
		return errors.New("invalid_type")
	}

	// 2. 6자리 인증코드 생성 (100000 ~ 999999)
	rand.Seed(time.Now().UnixNano())
	code := strconv.Itoa(rand.Intn(900000) + 100000)

	// 3. Redis에 저장 (Key: "auth:타입:이메일", Value: 코드, TTL: 3분)
	// 예: auth:signup:test@naver.com -> 123456
	redisKey := "auth:" + authType + ":" + emailAddr
	err = redis.SetData(redisKey, code, 3*time.Minute)
	if err != nil {
		return err
	}

	// 4. 이메일 전송
	err = email.SendVerificationCode(emailAddr, code)
	if err != nil {
		return err
	}

	return nil
}

// === 인증번호 검증 서비스 ===
// 반환값: (reset_token, error) -> signup일 땐 token이 빈 문자열입니다.
func VerifyEmailCode(emailAddr string, inputCode string, authType string) (string, error) {
	// 1. Redis에서 저장된 코드 조회
	redisKey := "auth:" + authType + ":" + emailAddr
	storedCode, err := redis.GetData(redisKey)

	// 코드가 없으면 (만료되었거나 키가 없음)
	if err != nil {
		return "", errors.New("expired_or_invalid")
	}

	// 2. 코드 비교
	if storedCode != inputCode {
		return "", errors.New("mismatch")
	}

	// 3. 인증 성공 후 처리 (인증번호는 재사용 못하게 삭제)
	redis.DeleteData(redisKey)

	// 4. 타입별 분기 처리
	if authType == "signup" {
		// 회원가입용은 단순히 성공 여부만 중요하므로 여기서 끝
		// (보안 강화: 실제로는 "verified:email" 같은 키를 Redis에 저장해서 회원가입 API가 체크하게 하면 더 좋습니다)
		return "", nil
	} else if authType == "reset" {
		// 비밀번호 찾기용은 '토큰'을 발급해야 함
		token, err := utils.GenerateRandomToken()
		if err != nil {
			return "", err
		}

		// [중요] 토큰 저장 (Key: "reset_token:토큰값", Value: "이메일", TTL: 10분)
		// 나중에 비밀번호 변경 API에서 이 토큰을 받으면, 누구의 이메일인지 역추적하기 위함입니다.
		tokenKey := "reset_token:" + token
		err = redis.SetData(tokenKey, emailAddr, 10*time.Minute)
		if err != nil {
			return "", err
		}

		return token, nil
	}

	return "", errors.New("invalid_type")
}

// === 비밀번호 재설정 서비스 ===
func ResetPassword(emailAddr string, resetToken string, newPassword string) error {
	// 1. Redis에서 토큰 검증
	// 저장할 때 Key: "reset_token:토큰값", Value: "이메일" 로 저장했었습니다.
	redisKey := "reset_token:" + resetToken
	storedEmail, err := redis.GetData(redisKey)

	// 토큰이 만료되었거나 없을 때
	if err != nil {
		return errors.New("invalid_token")
	}

	// 2. 토큰의 주인(이메일)과 요청한 이메일이 일치하는지 확인
	// (다른 사람의 이메일로 비밀번호를 바꾸려는 시도 차단)
	if storedEmail != emailAddr {
		return errors.New("email_mismatch")
	}

	// 3. 유저 조회
	user, err := repositories.FindUserByUsername(emailAddr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user_not_found")
		}
		return err
	}

	// 4. 새 비밀번호 해싱
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 5. DB 업데이트
	err = repositories.UpdateUserFields(&user, map[string]interface{}{
		"password_hash": hashedPassword,
	})

	if err != nil {
		return err
	}

	// 6. 사용한 토큰 삭제 (재사용 방지)
	redis.DeleteData(redisKey)

	return nil
}
