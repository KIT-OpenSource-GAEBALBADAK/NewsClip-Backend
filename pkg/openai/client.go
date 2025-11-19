package openai

import (
	"context"
	"encoding/json" // [신규] JSON 파싱을 위해 추가
	"fmt"
	"newsclip/backend/config"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// [신규] AI 응답을 담을 구조체
type AiShortResponse struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// [수정] 반환값 변경: (string) -> (title, summary, error)
func SummarizeNews(content string) (string, string, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	if len(content) > 3000 {
		content = content[:3000]
	}

	// [수정] 프롬프트: JSON 포맷을 강제함
	prompt := fmt.Sprintf(`
	다음 뉴스 기사를 읽고 '쇼츠(Shorts)'용 제목과 요약을 작성해줘.

	[요구사항]
	1. 제목: 20자 이내로 핵심을 찌르는 흥미로운 문구.
	2. 요약: 2줄 이내, 핵심 내용만, 말투는 뉴스이다보니 정중한 말투로.
	3. 응답 형식: 반드시 아래 JSON 포맷만 반환해 (마크다운 코드블럭 사용 금지).

	예시)
	"title":젊은 세대를 위한 새로운 뉴스 소비 패턴
	"summary":최근 연구에 따르면 Z세대와 밀레니얼 세대는 전통적인
	뉴스 매체보다 소셜 미디어와 모바일 플랫폼을 통해
	뉴스를 소비하는 경향이 강해지고있습니다.
	
	{"title": "여기에 제목", "summary": "여기에 요약 내용"}

	[기사 내용]
	%s
	`, content)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// (선택) JSON 모드 활성화 (최신 모델에서 지원하지만, 프롬프트만으로도 충분히 잘 동작함)
			// ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJsonObject},
		},
	)

	if err != nil {
		return "", "", err
	}

	// 응답 문자열 (JSON)
	rawContent := resp.Choices[0].Message.Content

	// 혹시 모를 마크다운 백틱 제거 (```json ... ```)
	rawContent = strings.TrimPrefix(rawContent, "```json")
	rawContent = strings.TrimPrefix(rawContent, "```")
	rawContent = strings.TrimSuffix(rawContent, "```")
	rawContent = strings.TrimSpace(rawContent)

	// JSON 파싱
	var aiResult AiShortResponse
	if err := json.Unmarshal([]byte(rawContent), &aiResult); err != nil {
		return "", "", fmt.Errorf("AI 응답 파싱 실패: %v | 원본: %s", err, rawContent)
	}

	return aiResult.Title, aiResult.Summary, nil
}
