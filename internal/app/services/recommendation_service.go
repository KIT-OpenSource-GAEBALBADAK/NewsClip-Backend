package services

import (
	"math"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"sort"
	"time"
)

// 개별 추천 뉴스 DTO (API 명세서 A 타입 가정)
type RecommendedNewsItemDTO struct {
	NewsID      uint      `json:"newsId"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl"`
	PublishedAt time.Time `json:"publishedAt"`
}

// 최종 응답 DTO
type RecommendedNewsResponseDTO struct {
	News []RecommendedNewsItemDTO `json:"news"`
}

// 사용자 선호 기반 뉴스 추천 서비스
func GetRecommendedNews(userID uint, size int) (*RecommendedNewsResponseDTO, error) {

	// 기본 추천 개수 = 5개
	if size <= 0 {
		size = 5
	}

	// ===== 1. 선호 카테고리(P1) =====
	preferredCategories, err := repositories.GetPreferredCategories(userID)
	if err != nil {
		return nil, err
	}
	preferredSet := make(map[string]bool, len(preferredCategories))
	for _, c := range preferredCategories {
		preferredSet[c] = true
	}

	// ===== 2. 북마크 기반 카테고리 통계(P2) =====
	bookmarkCounts, err := repositories.GetBookmarkCategoryStats(userID)
	if err != nil {
		return nil, err
	}

	// ===== 3. 상호작용(좋아요/싫어요) 기반 카테고리 통계(P3) =====
	likeCounts, dislikeCounts, err := repositories.GetInteractionCategoryStats(userID)
	if err != nil {
		return nil, err
	}

	// ===== 4. 추천 후보 뉴스 조회 =====
	candidateSize := size * 10
	if candidateSize < size {
		candidateSize = size
	}
	if candidateSize > 200 {
		candidateSize = 200
	}

	candidates, err := repositories.FindNewsCandidatesForRecommendation(userID, 60, candidateSize)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return &RecommendedNewsResponseDTO{News: []RecommendedNewsItemDTO{}}, nil
	}

	// ===== 5. 각 뉴스별 점수 계산 =====
	type scoredNews struct {
		News  models.News
		Score float64
	}

	scoredList := make([]scoredNews, 0, len(candidates))

	for _, news := range candidates {
		category := news.Category
		var score float64

		// P1: 선호 카테고리 +30
		if preferredSet[category] {
			score += 30.0
		}

		// P2: 북마크 로그 스케일
		if cnt, ok := bookmarkCounts[category]; ok && cnt > 0 {
			score += 5.0 * math.Log(float64(cnt)+1.0)
		}

		// P3: 좋아요/싫어요 로그 스케일
		if cnt, ok := likeCounts[category]; ok && cnt > 0 {
			score += 5.0 * math.Log(float64(cnt)+1.0)
		}
		if cnt, ok := dislikeCounts[category]; ok && cnt > 0 {
			score -= 5.0 * math.Log(float64(cnt)+1.0)
		}

		scoredList = append(scoredList, scoredNews{
			News:  news,
			Score: score,
		})
	}

	// ===== 6. 점수 순 정렬 (동점이면 최신 우선) =====
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].Score == scoredList[j].Score {
			return scoredList[i].News.PublishedAt.After(scoredList[j].News.PublishedAt)
		}
		return scoredList[i].Score > scoredList[j].Score
	})

	// ===== 7. 카테고리별 최대 N개(2개) 제한 적용 =====
	maxPerCategory := 2
	categoryCounter := make(map[string]int)

	finalList := make([]scoredNews, 0, size)

	for _, item := range scoredList {
		cat := item.News.Category

		// 해당 카테고리에 이미 2개 들어갔다면 스킵
		if categoryCounter[cat] >= maxPerCategory {
			continue
		}

		finalList = append(finalList, item)
		categoryCounter[cat]++

		// 정해진 size만큼 채우면 종료
		if len(finalList) >= size {
			break
		}
	}

	// 만약 부족하면(카테고리 제한 때문에), 다시 나머지에서 채움
	if len(finalList) < size {
		for _, item := range scoredList {
			alreadyIncluded := false
			for _, f := range finalList {
				if f.News.ID == item.News.ID {
					alreadyIncluded = true
					break
				}
			}
			if alreadyIncluded {
				continue
			}

			finalList = append(finalList, item)
			if len(finalList) >= size {
				break
			}
		}
	}

	// ===== 8. DTO 변환 =====
	items := make([]RecommendedNewsItemDTO, len(finalList))
	for i, sn := range finalList {
		n := sn.News
		items[i] = RecommendedNewsItemDTO{
			NewsID:      n.ID,
			Title:       n.Title,
			Source:      n.Source,
			Category:    n.Category,
			ImageURL:    n.ImageURL,
			PublishedAt: n.PublishedAt,
		}
	}

	return &RecommendedNewsResponseDTO{
		News: items,
	}, nil
}
