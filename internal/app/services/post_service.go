package services

import (
	"newsclip/backend/internal/app/repositories"
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
			CreatedAt:    post.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
