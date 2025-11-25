package repositories

import (
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
)

func GetPreferredCategories(userID uint) ([]string, error) {
	var list []models.UserPreferredCategory
	err := config.DB.Where("user_id = ?", userID).Find(&list).Error
	if err != nil {
		return nil, err
	}

	result := make([]string, len(list))
	for i, item := range list {
		result[i] = item.CategoryName
	}

	return result, nil
}
