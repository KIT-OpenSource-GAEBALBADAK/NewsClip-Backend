package services

import (
	"errors"
	"html"
	"log"
	"net/http"
	"net/url"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/repositories"
	"newsclip/backend/pkg/navernews"
	"regexp"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

// === HTML 태그를 제거하기 위한 정규식 컴파일러 ===
// (<...> 형태의 모든 태그를 찾음, 서버 시작 시 1회만 컴파일)
var tagStripper = regexp.MustCompile("<[^>]*>")

// === 문자열을 정리하는 헬퍼 함수 ===
func cleanString(s string) string {
	// 1. HTML 엔티티 디코딩 (예: &quot; -> ", &lt; -> <)
	unescaped := html.UnescapeString(s)

	// 2. HTML 태그 제거 (예: <b>...</b> -> ...)
	stripped := tagStripper.ReplaceAllString(unescaped, "")

	return stripped
}

// === 함수명 변경 및 기능 확장 (og:image + og:site_name) ===
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
// === FetchAllCategories ===
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

// === FetchAndStoreNews 함수 ===
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

// === 뉴스 목록 조회 서비스 ===
// (지금은 레포지토리를 호출만 하지만, 추후 'isBookmarked' 로직이 여기에 추가됨)
// (DTO를 사용하여 API 응답 구조를 정의)
type NewsListDTO struct {
	News       []models.News `json:"news"`
	TotalItems int64         `json:"totalItems"`
	TotalPages int           `json:"totalPages"`
}

// === 목록 응답용 DTO 정의 ===
type NewsResponseDTO struct {
	models.News       // 기존 뉴스 필드 모두 포함
	IsLiked      bool `json:"is_liked"`
	IsDisliked   bool `json:"is_disliked"`
	IsBookmarked bool `json:"is_bookmarked"`
}

type NewsListResponseDTO struct {
	News       []NewsResponseDTO `json:"news"`
	TotalItems int64             `json:"totalItems"`
	TotalPages int               `json:"totalPages"`
}

// === 뉴스 목록 조회 서비스 ===
func GetNewsList(category string, page int, size int, userID uint) (*NewsListResponseDTO, error) {

	// 1. 뉴스 데이터 조회
	newsList, totalCount, totalPages, err := repositories.GetNewsByCategory(category, page, size)
	if err != nil {
		return nil, err
	}

	// 2. 뉴스 ID 추출
	newsIDs := make([]uint, len(newsList))
	for i, news := range newsList {
		newsIDs[i] = news.ID
	}

	// 3. 상호작용 및 북마크 상태 조회 (로그인 유저인 경우만)
	// Map을 사용하여 O(1)로 조회 속도 최적화
	interactionMap := make(map[uint]string) // newsID -> "like" or "dislike"
	bookmarkMap := make(map[uint]bool)      // newsID -> true

	if userID != 0 && len(newsIDs) > 0 {
		// 3-1. 좋아요/싫어요 조회
		interactions, _ := repositories.GetNewsInteractionsByIDs(userID, newsIDs)
		for _, inter := range interactions {
			interactionMap[inter.NewsID] = inter.InteractionType
		}

		// 3-2. 북마크 조회
		bookmarks, _ := repositories.GetNewsBookmarksByIDs(userID, newsIDs)
		for _, bm := range bookmarks {
			bookmarkMap[bm.NewsID] = true
		}
	}

	// 4. DTO 변환 및 데이터 병합
	dtos := make([]NewsResponseDTO, len(newsList))
	for i, news := range newsList {
		// Map에서 상태 확인
		interType, hasInteraction := interactionMap[news.ID]
		isBookmarked := bookmarkMap[news.ID]

		dtos[i] = NewsResponseDTO{
			News:         news,
			IsLiked:      hasInteraction && interType == "like",
			IsDisliked:   hasInteraction && interType == "dislike",
			IsBookmarked: isBookmarked,
		}
	}

	// 5. 최종 응답 반환
	return &NewsListResponseDTO{
		News:       dtos,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}, nil
}

// === 뉴스 상세 조회 DTO (snake_case 적용) ===
type NewsDetailDTO struct {
	models.News
	IsBookmarked bool `json:"is_bookmarked"` // [수정] isBookmarked -> is_bookmarked
	IsLiked      bool `json:"is_liked"`      // [수정] isLiked -> is_liked
	IsDisliked   bool `json:"is_disliked"`   // [수정] isDisliked -> is_disliked
}

// === 뉴스 상세 조회 서비스 ===
func GetNewsDetail(newsID uint, userID uint) (*NewsDetailDTO, error) {

	// 1. (병렬 처리) DB에서 뉴스 정보 가져오기
	newsChan := make(chan models.News)
	errChan := make(chan error)

	go func() {
		news, err := repositories.FindNewsByID(newsID)
		if err != nil {
			errChan <- err
			return
		}
		newsChan <- news
	}()

	// 2. (백그라운드) 조회수 1 증가
	go func() {
		_ = repositories.IncrementNewsViewCount(newsID)
	}()

	// 3. [신규 로직] 사용자별 상호작용 정보 가져오기
	isBookmarked := false
	isLiked := false
	isDisliked := false

	// 로그인한 유저라면 DB에서 상태 조회
	if userID != 0 {
		// 3-1. 좋아요/싫어요 상태 확인
		// (FindNewsInteraction은 트랜잭션 객체(*gorm.DB)를 받으므로 config.DB를 전달)
		interaction, err := repositories.FindNewsInteraction(config.DB, userID, newsID)
		if err == nil {
			// 레코드가 존재하면 상태 설정
			isLiked = (interaction.InteractionType == "like")
			isDisliked = (interaction.InteractionType == "dislike")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// RecordNotFound 외의 에러는 로그를 찍거나 처리 (여기선 무시하고 false 유지)
		}

		// 3-2. 북마크 상태 확인
		_, err = repositories.FindBookmark(userID, newsID)
		if err == nil {
			// 에러가 없으면 북마크가 존재하는 것
			isBookmarked = true
		}
	}

	// 4. 뉴스 정보 로드 대기
	var news models.News
	select {
	case news = <-newsChan:
		// 성공
	case err := <-errChan:
		return nil, err
	}

	// 5. DTO 반환
	response := &NewsDetailDTO{
		News:         news,
		IsBookmarked: isBookmarked,
		IsLiked:      isLiked,
		IsDisliked:   isDisliked,
	}

	return response, nil
}

// === 상호작용 DTO ===
type InteractionRequest struct {
	InteractionType string `json:"interaction_type" binding:"required"`
}

type InteractionResponseDTO struct {
	IsLiked      bool `json:"isLiked"`
	IsDisliked   bool `json:"isDisliked"`
	LikeCount    int  `json:"likeCount"`
	DislikeCount int  `json:"dislikeCount"`
}

// === 뉴스 상호작용 서비스 ===
func InteractWithNews(userID, newsID uint, newType string) (*InteractionResponseDTO, error) {

	// 최종 응답으로 사용할 변수
	var finalResponse InteractionResponseDTO

	// 트랜잭션 시작
	err := config.DB.Transaction(func(tx *gorm.DB) error {

		// 1. 기존 상호작용 조회
		existingInteraction, err := repositories.FindNewsInteraction(tx, userID, newsID)

		var likeDelta, dislikeDelta int = 0, 0

		// --- 3가지 시나리오 분기 ---

		// [시나리오 1] 최초의 상호작용 (gorm.ErrRecordNotFound)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newInteraction := &models.NewsInteraction{
				UserID:          userID,
				NewsID:          newsID,
				InteractionType: newType,
			}
			if err := repositories.CreateNewsInteraction(tx, newInteraction); err != nil {
				return err
			}

			// 캐시 카운트 +1
			if newType == "like" {
				likeDelta = 1
			} else {
				dislikeDelta = 1
			}

			// 최종 상태 설정
			finalResponse.IsLiked = (newType == "like")
			finalResponse.IsDisliked = (newType == "dislike")
		} else if err == nil { // [시나리오 2] 이미 상호작용이 존재함
			// [2-A] 같은 버튼을 또 눌렀음 (취소)
			if existingInteraction.InteractionType == newType {
				if err := repositories.DeleteNewsInteraction(tx, &existingInteraction); err != nil {
					return err
				}
				// 캐시 카운트 -1
				if newType == "like" {
					likeDelta = -1
				} else {
					dislikeDelta = -1
				}

				// 최종 상태 설정 (둘 다 false)
				finalResponse.IsLiked = false
				finalResponse.IsDisliked = false
			} else { // [2-B] 다른 버튼을 눌렀음 (전환: like -> dislike)
				if err := repositories.UpdateNewsInteraction(tx, &existingInteraction, newType); err != nil {
					return err
				}
				// 캐시 카운트 전환 (예: like -1, dislike +1)
				if newType == "like" { // 'dislike' -> 'like'로 전환
					likeDelta = 1
					dislikeDelta = -1
				} else { // 'like' -> 'dislike'로 전환
					likeDelta = -1
					dislikeDelta = 1
				}

				// 최종 상태 설정
				finalResponse.IsLiked = (newType == "like")
				finalResponse.IsDisliked = (newType == "dislike")
			}

		} else { // [시나리오 3] 기타 DB 오류
			return err
		}

		// 2. 'news' 테이블의 캐시 카운트 업데이트
		if err := repositories.UpdateNewsCounts(tx, newsID, likeDelta, dislikeDelta); err != nil {
			return err
		}

		// 3. 최종 카운트를 DB에서 다시 읽어와서 응답에 담기
		var news models.News
		if err := tx.Select("like_count", "dislike_count").First(&news, newsID).Error; err != nil {
			return err
		}

		finalResponse.LikeCount = news.LikeCount
		finalResponse.DislikeCount = news.DislikeCount

		return nil // 트랜잭션 커밋
	}) // --- 트랜잭션 종료 ---

	if err != nil {
		return nil, err
	}

	return &finalResponse, nil
}

// === 뉴스 북마크 토글 서비스 ===
// (최종 북마크 상태를 bool로 반환)
func ToggleBookmark(userID, newsID uint) (bool, error) {

	// 1. 북마크가 이미 존재하는지 확인
	existingBookmark, err := repositories.FindBookmark(userID, newsID)

	// [시나리오 1] 북마크가 존재하지 않음 (gorm.ErrRecordNotFound)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newBookmark := &models.NewsBookmark{
			UserID: userID,
			NewsID: newsID,
		}

		if err := repositories.CreateBookmark(newBookmark); err != nil {
			// (참고: 만약 newsID가 존재하지 않아 FK 에러가 나면 여기서 걸림)
			return false, err // 생성 실패
		}

		// [수정] 생성에 성공했으므로 'true' (북마크 됨) 상태 반환
		return true, nil
	}

	// [시나리오 2] 북마크가 이미 존재함 (err == nil)
	if err == nil {
		if err := repositories.DeleteBookmark(&existingBookmark); err != nil {
			return false, err // 삭제 실패
		}

		// [수정] 삭제에 성공했으므로 'false' (북마크 취소됨) 상태 반환
		return false, nil
	}

	// [시나리오 3] 기타 DB 오류
	return false, err
}

// === 북마크 목록 조회 DTO ===

// BookmarkedNewsItemDTO: API 응답용 개별 뉴스 DTO (is_bookmarked 추가)
// (models.News를 임베딩하여 모든 필드를 상속받음)
type BookmarkedNewsItemDTO struct {
	models.News       // News 모델의 모든 필드 (ID, Title, Content...)
	IsBookmarked bool `json:"is_bookmarked"`
}

// BookmarkListResponseDTO: 최종 API 응답 DTO (페이지네이션 메타 포함)
type BookmarkListResponseDTO struct {
	News       []BookmarkedNewsItemDTO `json:"news"`
	TotalItems int64                   `json:"total_items"`
	TotalPages int                     `json:"total_pages"`
	Page       int                     `json:"page"`
	Size       int                     `json:"size"`
}

// === 북마크 목록 조회 서비스 ===
func GetBookmarkedNewsList(userID uint, page int, size int) (*BookmarkListResponseDTO, error) {

	// 1. 레포지토리에서 데이터 조회
	newsList, totalCount, totalPages, err := repositories.GetBookmarkedNews(userID, page, size)
	if err != nil {
		return nil, err
	}

	// 2. models.News -> BookmarkedNewsItemDTO로 변환
	// (이 목록은 북마크된 목록이므로 is_bookmarked는 항상 true)
	bookmarkedItems := make([]BookmarkedNewsItemDTO, len(newsList))
	for i, news := range newsList {
		bookmarkedItems[i] = BookmarkedNewsItemDTO{
			News:         news,
			IsBookmarked: true,
		}
	}

	// 3. 최종 응답 DTO 구성
	response := &BookmarkListResponseDTO{
		News:       bookmarkedItems,
		TotalItems: totalCount,
		TotalPages: totalPages,
		Page:       page,
		Size:       size,
	}

	return response, nil
}
