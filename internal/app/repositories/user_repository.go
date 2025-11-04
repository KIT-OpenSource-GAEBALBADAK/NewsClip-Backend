package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

// username으로 유저 조회
func FindUserByUsername(username string) (models.User, error) {
	var user models.User
	// GORM은 포인터 타입의 필드도 "username = ?"으로 조회 가능
	result := config.DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

// Provider와 ProviderID로 유저를 찾습니다.
func FindUserBySocial(provider, providerID string) (models.User, error) {
	var user models.User
	result := config.DB.Where("provider = ? AND provider_id = ?", provider, &providerID).First(&user)
	return user, result.Error
}

// userID로 유저 조회 (Refresh 및 프로필 설정 시 필요)
func FindUserByID(id uint) (models.User, error) {
	var user models.User
	result := config.DB.First(&user, id)
	return user, result.Error
}

// 유저 생성
func CreateUser(user *models.User) error {
	return config.DB.Create(user).Error
}

// [신규] 유저 프로필 업데이트 (닉네임, 프로필 이미지)
func UpdateUserProfile(user *models.User, nickname, profileImage string) error {
	result := config.DB.Model(user).Updates(models.User{
		Nickname:     &nickname,
		ProfileImage: &profileImage,
	})
	return result.Error
}

// 유저 통계 포함 조회 (posts, comments, likes 수)
func GetUserProfile(userID uint) (map[string]interface{}, error) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var postCount int64
	var commentCount int64
	var likeCount int64

	config.DB.Model(&models.Post{}).Where("user_id = ?", userID).Count(&postCount)
	config.DB.Model(&models.PostComment{}).Where("user_id = ?", userID).Count(&commentCount)
	config.DB.Model(&models.PostLike{}).Where("user_id = ?", userID).Count(&likeCount)

	data := map[string]interface{}{
		"nickname": user.Nickname,
		"role":     user.Role,
		"stats": map[string]interface{}{
			"postCount":    postCount,
			"commentCount": commentCount,
			"likeCount":    likeCount,
		},
		"profile_image": func() string {
			if user.ProfileImage == nil || *user.ProfileImage == "" {
				return "https://newsclip.duckdns.org/v1/images/default_profile.png"
			}
			return *user.ProfileImage
		}(),
	}

	return data, nil

}
