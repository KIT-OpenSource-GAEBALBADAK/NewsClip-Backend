package services

import (
	"errors"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"time"

	"gorm.io/gorm"
)

type CommunityPostDTO struct {
	PostID       uint      `json:"postId"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Category     string    `json:"category"`
	Author       AuthorDTO `json:"author"`
	Images       []string  `json:"images"`
	CreatedAt    string    `json:"createdAt"`
	ViewCount    int       `json:"viewCount"`
	LikeCount    int       `json:"likeCount"`
	DislikeCount int       `json:"dislikeCount"`
	CommentCount int       `json:"commentCount"`
}

type AuthorDTO struct {
	Nickname     *string `json:"nickname"`
	ProfileImage string  `json:"profileImage"`
	Role         string  `json:"role"`
}

type CommunityPostListResponse struct {
	Posts []CommunityPostDTO `json:"posts"`
}

func GetCommunityPosts(postType string, page, size int) (*CommunityPostListResponse, error) {

	posts, err := repositories.GetPostsWithRelations(postType, page, size)
	if err != nil {
		return nil, err
	}

	var result []CommunityPostDTO

	for _, post := range posts {

		profileImage := "https://newsclip.duckdns.org/v1/images/default_profile.png"
		if post.User.ProfileImage != nil && *post.User.ProfileImage != "" {
			profileImage = *post.User.ProfileImage
		}

		imageURLs := make([]string, len(post.Images))
		for i, img := range post.Images {
			imageURLs[i] = img.ImageURL
		}

		result = append(result, CommunityPostDTO{
			PostID:   post.ID,
			Title:    post.Title,
			Content:  post.Content,
			Category: post.Category,
			Author: AuthorDTO{
				Nickname:     post.User.Nickname,
				ProfileImage: profileImage,
				Role:         post.User.Role,
			},
			Images:       imageURLs,
			CreatedAt:    post.CreatedAt.Format(time.RFC3339),
			ViewCount:    post.ViewCount,
			LikeCount:    post.LikeCount,
			DislikeCount: post.DislikeCount,
			CommentCount: post.CommentCount,
		})
	}

	return &CommunityPostListResponse{
		Posts: result,
	}, nil
}

func CreatePost(userID uint, title, content, category string, imageURLs []string) (*models.Post, error) {

	post := models.Post{
		UserID:   userID,
		Title:    title,
		Content:  content,
		Category: category,
		Section:  "general",
	}

	err := repositories.CreatePost(&post)
	if err != nil {
		return nil, err
	}

	// 이미지 저장
	for _, img := range imageURLs {
		repositories.CreatePostImage(post.ID, img)
	}

	return &post, nil
}

// === 게시글 상호작용 응답 DTO ===
type PostInteractionResponseDTO struct {
	IsLiked      bool `json:"is_liked"`
	IsDisliked   bool `json:"is_disliked"`
	LikeCount    int  `json:"like_count"`
	DislikeCount int  `json:"dislike_count"`
}

// === 게시글 상호작용 서비스 ===
func InteractWithPost(userID, postID uint, newType string) (*PostInteractionResponseDTO, error) {

	var finalResponse PostInteractionResponseDTO

	err := config.DB.Transaction(func(tx *gorm.DB) error {

		// 1. 기존 상호작용 조회
		existingInteraction, err := repositories.FindPostInteraction(tx, userID, postID)

		var likeDelta, dislikeDelta int = 0, 0

		// [시나리오 1] 최초 상호작용
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newInteraction := &models.PostInteraction{
				UserID:          userID,
				PostID:          postID,
				InteractionType: newType,
			}
			if err := repositories.CreatePostInteraction(tx, newInteraction); err != nil {
				return err
			}

			if newType == "like" {
				likeDelta = 1
			} else {
				dislikeDelta = 1
			}

			finalResponse.IsLiked = (newType == "like")
			finalResponse.IsDisliked = (newType == "dislike")

			// [시나리오 2] 이미 존재함
		} else if err == nil {
			// [2-A] 취소 (같은 타입 클릭)
			if existingInteraction.InteractionType == newType {
				if err := repositories.DeletePostInteraction(tx, &existingInteraction); err != nil {
					return err
				}
				if newType == "like" {
					likeDelta = -1
				} else {
					dislikeDelta = -1
				}

				finalResponse.IsLiked = false
				finalResponse.IsDisliked = false
			} else {
				// [2-B] 전환 (다른 타입 클릭)
				if err := repositories.UpdatePostInteraction(tx, &existingInteraction, newType); err != nil {
					return err
				}
				if newType == "like" { // dislike -> like
					likeDelta = 1
					dislikeDelta = -1
				} else { // like -> dislike
					likeDelta = -1
					dislikeDelta = 1
				}

				finalResponse.IsLiked = (newType == "like")
				finalResponse.IsDisliked = (newType == "dislike")
			}
		} else {
			return err
		}

		// 2. 카운트 업데이트
		if err := repositories.UpdatePostCounts(tx, postID, likeDelta, dislikeDelta); err != nil {
			return err
		}

		// 3. 최신 카운트 조회
		var post models.Post
		if err := tx.Select("like_count", "dislike_count").First(&post, postID).Error; err != nil {
			return err
		}

		finalResponse.LikeCount = post.LikeCount
		finalResponse.DislikeCount = post.DislikeCount

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &finalResponse, nil
}

// === [5.4] 내가 쓴 게시글 삭제 ===
func DeleteMyPost(userID, postID uint) error {

	// 1. 게시글 조회
	post, err := repositories.FindPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("게시글을 찾을 수 없습니다")
		}
		return err
	}

	// 2. 작성자 검증
	if post.UserID != userID {
		return errors.New("본인이 작성한 게시글만 삭제할 수 있습니다")
	}

	// 3. 게시글 삭제
	if err := repositories.DeletePost(&post); err != nil {
		return err
	}

	return nil
}
