package services

import "newsclip/backend/internal/app/repositories"

func GetMyProfile(userID uint) (map[string]interface{}, error) {
	return repositories.GetUserProfile(userID)
}
