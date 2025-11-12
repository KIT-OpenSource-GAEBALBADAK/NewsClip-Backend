package services

import (
	"html"
	"log"
	"net/http"
	"net/url"
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

// === [수정] 함수명 변경 및 기능 확장 (og:image + og:site_name) ===
// (url) -> (imageURL, siteName, error)
func getPageMetadata(url string) (string, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", log.Output(2, "request failed")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	var imageURL, siteName string

	// meta 태그를 한 번만 순회하며 두 가지 정보를 찾습니다.
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")

		if property == "og:image" {
			imageURL, _ = s.Attr("content")
		}

		if property == "og:site_name" {
			siteName, _ = s.Attr("content")
		}
	})

	return imageURL, siteName, nil
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
// (언론사명, 작성시간 추가)
func FetchAndStoreNews(query string, display int) error {
	client := navernews.NewClient()

	resp, err := client.SearchNews(query, display, 1)
	if err != nil {
		return err
	}

	log.Printf("Fetched %d items for query '%s' from Naver.", len(resp.Items), query)

	var newsToCreate []models.News
	for _, item := range resp.Items {

		externalID := item.Link

		_, err := repositories.FindNewsByExternalID(externalID)
		if err == nil {
			continue // 중복
		}

		// --- [수정] 1. 메타데이터(이미지, 언론사) 가져오기 ---
		imageURL, publisherName, err := getPageMetadata(item.Originallink)
		if err != nil {
			log.Printf("Failed to get metadata for %s: %v", item.Title, err)
			imageURL = ""
		}

		// [수정] 1-1. 언론사명이 비어있을 경우, 원문 링크의 호스트(도메인)로 대체
		if publisherName == "" {
			parsedURL, err := url.Parse(item.Originallink)
			if err == nil {
				publisherName = parsedURL.Host // 예: "www.yna.co.kr"
			} else {
				publisherName = "Unknown" // 파싱 실패 시
			}
		}

		// --- [수정] 2. 원본 기사 작성 시간(pubDate) 파싱 ---
		// Naver API의 pubDate는 "RFC 1123Z" 형식 (예: Mon, 10 Nov 2025 14:30:00 +0900)
		pubTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// [수정] 3. 파싱 실패 시(요구사항 #3) 현재 시간으로 대체
			log.Printf("Failed to parse pubDate '%s', using current time. Error: %v", item.PubDate, err)
			pubTime = time.Now()
		}

		// --- [수정] 4. 문자열 정리 ---
		cleanTitle := cleanString(item.Title)
		cleanDescription := cleanString(item.Description)

		newsToCreate = append(newsToCreate, models.News{
			ExternalID:  externalID,
			Title:       cleanTitle,
			Content:     cleanDescription,
			Source:      publisherName, // [수정] "연합뉴스" 또는 "www.yna.co.kr"
			URL:         item.Link,
			Category:    query,
			ImageURL:    imageURL,
			PublishedAt: pubTime, // [신규] 원본 기사 시간
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

// === [신규] 뉴스 목록 조회 서비스 ===
// (지금은 레포지토리를 호출만 하지만, 추후 'isBookmarked' 로직이 여기에 추가됨)
// (DTO를 사용하여 API 응답 구조를 정의)
type NewsListDTO struct {
	News       []models.News `json:"news"`
	TotalItems int64         `json:"totalItems"`
	TotalPages int           `json:"totalPages"`
}

func GetNewsList(category string, page int, size int, userID uint) (*NewsListDTO, error) {

	// 1. 레포지토리에서 데이터 조회
	newsList, totalCount, totalPages, err := repositories.GetNewsByCategory(category, page, size)
	if err != nil {
		return nil, err
	}

	// 2. [향후 로직 추가]
	// if userID > 0 {
	//    - newsList에서 newsID 목록 추출
	//    - repositories.FindBookmarkedNewsIDs(userID, newsIDs) 호출
	//    - DTO를 새로 정의하고(NewsItemDTO), newsList를 순회하며 'isBookmarked' 값을 채워넣기
	// }

	// 3. (현재) DTO에 담아 반환
	response := &NewsListDTO{
		News:       newsList,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	return response, nil
}
