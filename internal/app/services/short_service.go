package services

import (
	"log"
	"net/http"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/pkg/openai"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// === [내부 함수] 뉴스 본문 크롤링 ===
// === 뉴스 본문 크롤링 (네이버 뉴스 #dic_area 구조 대응) ===
func crawlNewsContent(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	// 1. 네이버 뉴스의 본문 영역 ID (#dic_area) 선택
	selection := doc.Find("#dic_area")

	// 만약 #dic_area가 없으면 상위 클래스나 일반적인 구조로 폴백(Fallback)
	if selection.Length() == 0 {
		selection = doc.Find("#newsct_article")
	}
	// 그래도 없으면 일반적인 article 태그 시도
	if selection.Length() == 0 {
		selection = doc.Find("article")
	}

	// 2. 불필요한 요소 제거 (이미지 캡션, 사진 영역 등)
	// 제공해주신 HTML의 <span class="end_photo_org"> 제거
	selection.Find(".end_photo_org").Remove()
	selection.Find(".img_desc").Remove() // 일반적인 이미지 설명 제거
	selection.Find("img").Remove()       // 이미지 태그 제거
	selection.Find("script").Remove()    // 스크립트 제거
	selection.Find("iframe").Remove()    // 동영상 등 제거

	// 3. <br> 태그를 공백으로 치환
	// (그냥 .Text()를 하면 "안녕하세요<br>반갑습니다"가 "안녕하세요반갑습니다"로 붙어버림)
	selection.Find("br").ReplaceWithHtml(" ")

	// 4. 텍스트 추출 및 공백 정리
	text := selection.Text()

	// strings.Fields는 연속된 공백(스페이스, 탭, 줄바꿈)을 하나로 합쳐줍니다.
	cleanText := strings.Join(strings.Fields(text), " ")

	return cleanText, nil
}

// === [핵심] 쇼츠 생성 및 저장 서비스 (카테고리별 1개) ===
func GenerateShorts() error {
	log.Println("🤖 [Shorts Generator] Starting to generate shorts per category...")

	// 1. 대상 카테고리 목록 정의
	categories := []string{
		"정치", "경제", "문화", "환경", "기술", "스포츠",
		"라이프스타일", "건강", "교육", "음식", "여행", "패션",
	}

	generatedCount := 0

	// 2. 카테고리별로 루프 실행
	for _, category := range categories {
		// 각 카테고리별로 후보 뉴스 3개를 가져옵니다. (최신순)
		// (1순위가 크롤링 실패할 경우 2순위를 하기 위함)
		var candidates []models.News

		// 쿼리 조건:
		// 1. 해당 카테고리
		// 2. 최근 24시간 이내 기사
		// 3. 이미 쇼츠가 만들어진 기사는 제외 (SubQuery)
		err := config.DB.
			Where("category = ?", category).
			Where("created_at > ?", time.Now().Add(-24*time.Hour)).
			Where("id NOT IN (SELECT news_id FROM shorts)").
			Order("published_at DESC"). // 최신 기사 우선 (PublishedAt 기준)
			Limit(3).                   // 후보 3개
			Find(&candidates).Error

		if err != nil {
			log.Printf("⚠️ DB Error fetching candidates for '%s': %v", category, err)
			continue
		}

		if len(candidates) == 0 {
			log.Printf("ℹ️ No suitable news found for category '%s'", category)
			continue
		}

		// 3. 후보 뉴스 중 하나를 성공할 때까지 시도
		for _, news := range candidates {
			// A. 본문 크롤링
			fullContent, err := crawlNewsContent(news.URL)
			if err != nil || len(fullContent) < 100 {
				log.Printf("   Skipping NewsID %d (%s): Crawl failed or too short.", news.ID, category)
				continue
			}

			// B. OpenAI 요약 [수정됨]
			// title, summary 두 개의 값을 받음
			title, summary, err := openai.SummarizeNews(fullContent)
			if err != nil {
				log.Printf("⚠️ OpenAI failed for NewsID %d: %v", news.ID, err)
				continue
			}

			// C. 쇼츠 DB 저장 [수정됨]
			newShort := models.Short{
				NewsID:    news.ID,
				Title:     title, // [신규] AI가 지은 제목 저장
				Summary:   summary,
				ImageURL:  news.ImageURL,
				CreatedAt: time.Now(),
			}

			if err := config.DB.Create(&newShort).Error; err != nil {
				log.Printf("⚠️ DB Save failed for NewsID %d: %v", news.ID, err)
			} else {
				log.Printf("✅ Short generated: [%s] %s", category, title)
				generatedCount++
				break
			}
		}
	}

	log.Printf("🤖 [Shorts Generator] Finished. Generated %d new shorts.", generatedCount)
	return nil
}

// === 쇼츠 피드 응답 DTO ===
type ShortFeedItemDTO struct {
	ShortID        uint   `json:"shortId"`
	OriginalNewsID uint   `json:"originalNewsId"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	ImageURL       string `json:"imageUrl"`
	LikeCount      int    `json:"likeCount"`
	DislikeCount   int    `json:"dislikeCount"`
	CommentCount   int    `json:"commentCount"` // (Comment 테이블 Count 로직은 생략, 현재 0)
	IsLiked        bool   `json:"isLiked"`
	IsDisliked     bool   `json:"isDisliked"`
}

// === 쇼츠 피드 조회 서비스 ===
func GetShortsFeed(size int, userID uint) ([]ShortFeedItemDTO, error) {
	// 1. 최신 쇼츠 목록 가져오기
	shorts, err := repositories.FindRecentShorts(size)
	if err != nil {
		return nil, err
	}

	// 쇼츠가 없으면 빈 배열 반환
	if len(shorts) == 0 {
		return []ShortFeedItemDTO{}, nil
	}

	// 2. (로그인 유저라면) 상호작용 정보 가져오기
	//    - 조회된 쇼츠들의 ID만 추출
	shortIDs := make([]uint, len(shorts))
	for i, s := range shorts {
		shortIDs[i] = s.ID
	}

	//    - interactionMap[shortID] = "like" or "dislike"
	interactionMap := make(map[uint]string)

	if userID != 0 {
		interactions, err := repositories.FindShortInteractionsByIDs(userID, shortIDs)
		if err == nil {
			for _, inter := range interactions {
				interactionMap[inter.ShortID] = inter.InteractionType
			}
		}
	}

	// 3. DTO 변환
	feed := make([]ShortFeedItemDTO, len(shorts))
	for i, s := range shorts {
		// 상호작용 상태 확인
		interType, exists := interactionMap[s.ID]

		feed[i] = ShortFeedItemDTO{
			ShortID:        s.ID,
			OriginalNewsID: s.NewsID,
			Title:          s.Title,
			Summary:        s.Summary,
			ImageURL:       s.ImageURL,
			LikeCount:      s.LikeCount,
			DislikeCount:   s.DislikeCount,
			// CommentCount: len(s.Comments), // 필요시 preload 또는 별도 카운트
			IsLiked:    exists && interType == "like",
			IsDisliked: exists && interType == "dislike",
		}
	}

	return feed, nil
}
