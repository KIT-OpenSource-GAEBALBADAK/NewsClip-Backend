package services

import (
	"html"
	"log"
	"net/http"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/pkg/navernews"
	"regexp"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// === [신규] HTML 태그를 제거하기 위한 정규식 컴파일러 ===
// (<...> 형태의 모든 태그를 찾음, 서버 시작 시 1회만 컴파일)
var tagStripper = regexp.MustCompile("<[^>]*>")

// === [신규] 문자열을 정리하는 헬퍼 함수 ===
func cleanString(s string) string {
	// 1. HTML 엔티티 디코딩 (예: &quot; -> ", &lt; -> <)
	unescaped := html.UnescapeString(s)

	// 2. HTML 태그 제거 (예: <b>...</b> -> ...)
	stripped := tagStripper.ReplaceAllString(unescaped, "")

	return stripped
}

// === 원문 URL에서 OG:IMAGE 태그를 추출하는 함수 ===
func getOgpImage(url string) (string, error) {
	// 1. HTTP GET 요청
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", log.Output(2, "request failed")
	}

	// 2. HTML 파싱
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	// 3. "og:image" 메타 태그 검색
	var imageURL string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		// "property" 속성이 "og:image"인 태그를 찾음
		if property, _ := s.Attr("property"); property == "og:image" {
			// "content" 속성(실제 URL)을 가져옴
			imageURL, _ = s.Attr("content")
		}
	})

	if imageURL == "" {
		return "", log.Output(2, "og:image not found")
	}

	return imageURL, nil
}

// === 모든 카테고리 뉴스를 병렬로 수집하는 함수 ===
// === [수정] FetchAllCategories ===
func FetchAllCategories() error {
	categories := []string{
		"정치", "경제", "문화", "환경", "기술", "스포츠",
		"라이프스타일", "건강", "교육", "음식", "여행", "패션",
	}

	// [수정] 10개 -> 5개
	displayPerCategory := 5

	log.Printf("[Scheduler] Starting fetch for %d categories (%d items each)...", len(categories), displayPerCategory)

	var wg sync.WaitGroup
	wg.Add(len(categories))

	for _, category := range categories {
		cat := category
		go func() {
			defer wg.Done()
			log.Printf("[Scheduler] ... fetching category: %s", cat)

			// displayPerCategory (5)를 전달
			err := FetchAndStoreNews(cat, displayPerCategory)

			if err != nil {
				log.Printf("🔥 [Scheduler] FAILED category %s: %v", cat, err)
			}
		}()
	}

	wg.Wait()
	log.Println("[Scheduler] All category fetching routines finished.")
	return nil
}

// === [수정] FetchAndStoreNews 함수 ===
// (cleanString 함수를 적용)
func FetchAndStoreNews(query string, display int) error {
	client := navernews.NewClient()

	// 1. 네이버 API에서 뉴스 검색
	resp, err := client.SearchNews(query, display, 1)
	if err != nil {
		return err
	}

	log.Printf("Fetched %d items for query '%s' from Naver.", len(resp.Items), query)

	// 2. DB 모델로 변환 (이미지 크롤링 추가)
	var newsToCreate []models.News
	for _, item := range resp.Items {

		externalID := item.Link

		_, err := repositories.FindNewsByExternalID(externalID)
		if err == nil {
			continue // 중복이면 건너뛰기
		}

		imageURL, err := getOgpImage(item.Originallink)
		if err != nil {
			imageURL = ""
		}

		// === [수정] 저장 전에 문자열 정리 ===
		cleanTitle := cleanString(item.Title)
		cleanDescription := cleanString(item.Description)
		// =================================

		newsToCreate = append(newsToCreate, models.News{
			ExternalID: externalID,
			Title:      cleanTitle,       // [수정]
			Content:    cleanDescription, // [수정]
			Source:     item.Originallink,
			URL:        item.Link,
			Category:   query,
			ImageURL:   imageURL,
		})
	}

	// 3. DB에 일괄 저장
	if len(newsToCreate) > 0 {
		err = repositories.CreateNewsBatch(newsToCreate)
		if err != nil {
			return err
		}
		log.Printf("✅ Successfully stored %d new items for '%s' in DB.", len(newsToCreate), query)
	} else {
		log.Printf("No new items to store for '%s'.", query)
	}

	return nil
}

// === 오래된 뉴스 삭제 서비스 ===
func CleanupOldNews() error {
	// 1. 기준 날짜(14일 전) 계산
	// (0년, 0개월, -14일)
	cutoffDate := time.Now().AddDate(0, 0, -14)

	log.Printf("[Cleaner] Deleting news older than %s", cutoffDate.Format("2006-01-02"))

	// 2. 레포지토리 호출
	count, err := repositories.DeleteNewsOlderThan(cutoffDate)
	if err != nil {
		log.Printf("🔥 [Cleaner] FAILED: %v", err)
		return err
	}

	log.Printf("✅ [Cleaner] Successfully deleted %d old news items.", count)
	return nil
}
