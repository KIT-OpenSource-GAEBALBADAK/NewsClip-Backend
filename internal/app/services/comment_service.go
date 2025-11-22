package services

import (
	"errors"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"time"
)

// === [DTO] API 응답용 구조체 ===
type UserSummaryDTO struct {
	ID           uint   `json:"id"`
	Nickname     string `json:"nickname"`
	ProfileImage string `json:"profile_image"`
	Role         string `json:"role"`
}

type CommentResponseDTO struct {
	CommentID uint           `json:"comment_id"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	User      UserSummaryDTO `json:"user"`
}

// === 댓글 작성 서비스 ===
func CreateComment(targetType string, targetID, userID uint, content string) (uint, error) {
	switch targetType {
	case "news":
		comment := models.NewsComment{NewsID: targetID, UserID: userID, Content: content}
		err := repositories.CreateNewsComment(&comment)
		return comment.ID, err
	case "short":
		comment := models.ShortComment{ShortID: targetID, UserID: userID, Content: content}
		err := repositories.CreateShortComment(&comment)
		return comment.ID, err
	case "post":
		comment := models.PostComment{PostID: targetID, UserID: userID, Content: content}
		err := repositories.CreatePostComment(&comment)
		return comment.ID, err
	default:
		return 0, errors.New("잘못된 대상 타입입니다")
	}
}

// === 댓글 목록 조회 서비스 ===
func GetComments(targetType string, targetID uint) ([]CommentResponseDTO, error) {
	var dtos []CommentResponseDTO

	switch targetType {
	case "news":
		comments, err := repositories.GetNewsComments(targetID)
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			dtos = append(dtos, convertToDTO(c.ID, c.Content, c.CreatedAt, c.User))
		}
	case "short":
		comments, err := repositories.GetShortComments(targetID)
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			dtos = append(dtos, convertToDTO(c.ID, c.Content, c.CreatedAt, c.User))
		}
	case "post":
		comments, err := repositories.GetPostComments(targetID)
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			dtos = append(dtos, convertToDTO(c.ID, c.Content, c.CreatedAt, c.User))
		}
	default:
		return nil, errors.New("잘못된 대상 타입입니다")
	}

	return dtos, nil
}

// (헬퍼 함수) 모델 데이터를 DTO로 변환
func convertToDTO(id uint, content string, createdAt time.Time, user models.User) CommentResponseDTO {
	// User 모델의 Nickname 등이 포인터일 경우 안전하게 처리
	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}

	profileImage := ""
	if user.ProfileImage != nil {
		profileImage = *user.ProfileImage
	}

	return CommentResponseDTO{
		CommentID: id,
		Content:   content,
		CreatedAt: createdAt,
		User: UserSummaryDTO{
			ID:           user.ID,
			Nickname:     nickname,
			ProfileImage: profileImage,
			Role:         user.Role,
		},
	}
}
