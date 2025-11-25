package services

import "newsclip/backend/internal/app/repositories"

func GetPreferredCategories(userID uint) ([]string, error) {
	return repositories.GetPreferredCategories(userID)
}
