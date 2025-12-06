package services

import (
	"errors"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
)

// === API 응답용 DTO 구조체 정의 ===
type UserProfileDTO struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	ProfileImage string `json:"profile_image"`
	Role         string `json:"role"`
}

type UserStatsDTO struct {
	PostCount          int64 `json:"post_count"`
	CommentCount       int64 `json:"comment_count"`
	TotalReceivedLikes int64 `json:"total_received_likes"`
}

type MyProfileResponseDTO struct {
	User  UserProfileDTO `json:"user"`
	Stats UserStatsDTO   `json:"stats"`
}

// === [수정] 내 프로필 조회 (통계 포함) ===
// 기존 map[string]interface{} 반환에서 *MyProfileResponseDTO 반환으로 변경
func GetMyProfile(userID uint) (*MyProfileResponseDTO, error) {
	// 1. 사용자 기본 정보 조회
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 2. 사용자 활동 통계 조회 (repositories/user_repository.go에 추가한 함수 호출)
	stats, err := repositories.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	// 3. DTO 변환 (Pointer Safety Check)
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	profileImage := ""
	if user.ProfileImage != nil {
		profileImage = *user.ProfileImage
	}

	response := &MyProfileResponseDTO{
		User: UserProfileDTO{
			ID:           user.ID,
			Username:     username,
			Nickname:     nickname,
			ProfileImage: profileImage,
			Role:         user.Role,
		},
		Stats: UserStatsDTO{
			PostCount:          stats.PostCount,
			CommentCount:       stats.CommentCount,
			TotalReceivedLikes: stats.TotalReceivedLikes,
		},
	}

	return response, nil
}

// === [기존 유지] 프로필 수정 ===
func UpdateProfile(userID uint, nickname string, profileImage *string) (models.User, error) {
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return user, errors.New("유저를 찾을 수 없습니다")
	}

	updates := map[string]interface{}{}

	if nickname != "" {
		updates["nickname"] = nickname
	}

	if profileImage != nil {
		updates["profile_image"] = *profileImage
	}

	if len(updates) == 0 {
		return user, errors.New("변경할 정보가 없습니다")
	}

	// (참고: repositories에 UpdateUserFields 함수가 구현되어 있어야 합니다)
	err = repositories.UpdateUserFields(&user, updates)
	if err != nil {
		return user, errors.New("프로필 업데이트에 실패했습니다")
	}

	// 메모리 반영
	if nickname != "" {
		user.Nickname = &nickname
	}
	if profileImage != nil {
		user.ProfileImage = profileImage
	}

	return user, nil
}
