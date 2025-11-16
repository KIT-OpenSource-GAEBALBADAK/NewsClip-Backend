package services

import (
	"errors"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
)

func GetMyProfile(userID uint) (map[string]interface{}, error) {
	return repositories.GetUserProfile(userID)
}

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
