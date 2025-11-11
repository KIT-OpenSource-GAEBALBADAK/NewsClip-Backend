package navernews

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"newsclip/backend/config"
	"time"
)

const (
	apiURL = "https://openapi.naver.com/v1/search/news.json"
)

// Naver API에서 반환되는 전체 응답 구조체
type NaverNewsResponse struct {
	LastBuildDate string     `json:"lastBuildDate"`
	Total         int        `json:"total"`
	Start         int        `json:"start"`
	Display       int        `json:"display"`
	Items         []NewsItem `json:"items"`
}

// 개별 뉴스 아이템 구조체
type NewsItem struct {
	Title        string `json:"title"`
	Originallink string `json:"originallink"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	PubDate      string `json:"pubDate"` // (RFC 1123 format)
}

// API 클라이언트 구조체
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// 새 클라이언트 생성
func NewClient() *Client {
	return &Client{
		clientID:     config.GetEnv("NAVER_CLIENT_ID"),
		clientSecret: config.GetEnv("NAVER_CLIENT_SECRET"),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// 뉴스를 검색하는 함수
func (c *Client) SearchNews(query string, display int, start int) (*NaverNewsResponse, error) {
	// 1. 요청 URL 및 쿼리 파라미터 설정
	baseURL, _ := url.Parse(apiURL)
	params := url.Values{}
	params.Add("query", query)
	params.Add("display", fmt.Sprintf("%d", display))
	params.Add("start", fmt.Sprintf("%d", start))
	params.Add("sort", "sim") // sim (유사도순), date (날짜순)
	baseURL.RawQuery = params.Encode()

	// 2. HTTP 요청 생성
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// 3. (중요) Naver API는 헤더에 인증 정보를 담아 보냅니다.
	req.Header.Set("X-Naver-Client-Id", c.clientID)
	req.Header.Set("X-Naver-Client-Secret", c.clientSecret)

	// 4. API 요청
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("네이버 API 요청 실패: %s", resp.Status)
	}

	// 5. 응답 JSON 파싱
	var response NaverNewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
