package services

import "newsclip/backend/internal/app/repositories"

func GetPreferredCategories(userID uint) ([]string, error) {
	return repositories.GetPreferredCategories(userID)
}

func SetPreferredCategories(userID uint, categories []string) error {
	// 기존 목록 삭제
	err := repositories.ClearPreferredCategories(userID)
	if err != nil {
		return err
	}

	// 새 목록 저장
	for _, c := range categories {
		if err := repositories.AddPreferredCategory(userID, c); err != nil {
			return err
		}
	}

	return nil
}
