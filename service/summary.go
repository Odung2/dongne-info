package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"dongne-info/model"
	"dongne-info/repository"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// SummaryService 는 Claude API를 사용하여 행정 공고를 쉬운 말로 요약하는 서비스.
//
// 행정 용어로 작성된 공고의 사업명/분류/위치 정보를 Claude API에 전달하여
// 일반 주민이 이해할 수 있는 쉬운 설명, 진행 단계, 관련 사례를 생성한다.
//
// 핵심 원칙:
//   - 예측하지 않는다. "집값이 오른다/내린다" 같은 예측 절대 금지.
//   - 팩트만 요약한다. 행정 용어를 쉬운 말로 바꾸는 것이 목적.
//   - 진행 단계를 안내한다. 전체 절차 중 어디인지 알려준다.
//   - 모르면 모른다고 한다. 관련 사례를 지어내지 않는다.
//
// 사용 흐름:
//  1. 크롤러가 새 공고를 DB에 저장 (summary=NULL 상태)
//  2. SummarizeUnsummarized()가 미요약 공고를 조회
//  3. Claude API로 요약 생성
//  4. DB에 summary, stage, related 업데이트
//
// 사용 예:
//
//	svc := service.NewSummaryService(apiKey, repo)
//	count, err := svc.SummarizeUnsummarized(ctx, 10)
type SummaryService struct {
	client *anthropic.Client
	repo   *repository.AnnouncementRepository
}

// NewSummaryService 는 SummaryService를 생성한다.
//
// apiKey가 빈 문자열이면 nil을 반환한다.
// 이 경우 요약 기능은 비활성화되며, 크롤러는 요약 없이 공고만 저장한다.
func NewSummaryService(apiKey string, repo *repository.AnnouncementRepository) *SummaryService {
	if apiKey == "" {
		return nil
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &SummaryService{client: &client, repo: repo}
}

// summaryResponse 는 Claude API 응답을 파싱하기 위한 내부 구조체.
type summaryResponse struct {
	Summary string `json:"summary"`
	Stage   string `json:"stage"`
	Related string `json:"related"`
}

// SummarizeUnsummarized 는 아직 AI 요약이 없는 공고들을 찾아 일괄 요약한다.
//
// 크롤링 후 또는 별도 배치 작업으로 호출한다.
// 한 번에 최대 limit건을 처리하며, 개별 공고 요약 실패 시에도 나머지는 계속 진행한다.
func (s *SummaryService) SummarizeUnsummarized(ctx context.Context, limit int) (int, error) {
	announcements, err := s.repo.FindUnsummarized(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("미요약 공고 조회 실패: %w", err)
	}

	if len(announcements) == 0 {
		slog.Info("요약할 공고 없음")
		return 0, nil
	}

	slog.Info("요약 시작", "count", len(announcements))

	summarized := 0
	for _, a := range announcements {
		result, err := s.summarize(ctx, &a)
		if err != nil {
			slog.Error("요약 생성 실패", "id", a.ID, "title", a.Title, "error", err)
			continue
		}

		if err := s.repo.UpdateSummary(ctx, a.ID, result.Summary, result.Stage, result.Related); err != nil {
			slog.Error("요약 저장 실패", "id", a.ID, "error", err)
			continue
		}

		summarized++
		slog.Info("요약 완료", "id", a.ID, "district", a.District, "summary", result.Summary)
	}

	return summarized, nil
}

// summarize 는 단일 공고에 대해 Claude API를 호출하여 요약을 생성한다.
//
// 입력 데이터 출처 (서울시 upisRebuild API):
//   - Title: RGN_NM 사업명 (예: "은마아파트 재건축 정비구역")
//   - Type: SCLSF에서 분류 (예: "재건축")
//   - Action: RPT_TYPE 원본 (예: "신설")
//   - RawCategory: LCLSF > MCLSF > SCLSF (예: "의제처리구역 > 정비구역 > 재건축사업구역")
//   - Location: PSTN_NM (예: "강남구 대치동 316번지 일대")
func (s *SummaryService) summarize(ctx context.Context, a *model.Announcement) (*summaryResponse, error) {
	location := ""
	if a.Location != nil {
		location = *a.Location
	}
	rawCategory := ""
	if a.RawCategory != nil {
		rawCategory = *a.RawCategory
	}

	systemPrompt := "서울시 행정 공고를 동네 주민이 바로 이해할 수 있게 설명해줘.\n\n" +
		"규칙:\n" +
		"1. 해요체 사용 (\"~됐어요\", \"~이에요\").\n" +
		"2. 2~3문장으로 핵심만. 사업명과 위치를 꼭 포함해.\n" +
		"3. 행정 용어는 쉬운 말로. \"정비구역 지정\" → \"재건축을 공식적으로 시작할 수 있게 된 것\".\n" +
		"4. 집값 예측 절대 금지. 팩트만 설명.\n" +
		"5. 재개발/재건축 7단계(①기본계획 ②정비구역지정 ③조합설립 ④사업시행인가 ⑤관리처분인가 ⑥착공 ⑦입주) 중 판별 가능하면 표시. 판별 불가능하면 \"확인 필요\"라고 써.\n" +
		"6. related는 같은 지역/비슷한 유형의 사례를 알면 쓰고, 모르면 빈 문자열 \"\". 지어내지 마.\n\n" +
		"반드시 순수 JSON만 응답. 마크다운 코드블록이나 추가 설명 금지.\n" +
		"{\"summary\":\"쉬운 설명 2~3문장\",\"stage\":\"N/7단계 - 단계명 또는 확인 필요\",\"related\":\"관련 사례 또는 빈 문자열\"}"

	areaBefore := ""
	if a.AreaBefore != nil && *a.AreaBefore != "0" {
		areaBefore = *a.AreaBefore
	}
	areaAfter := ""
	if a.AreaAfter != nil && *a.AreaAfter != "0" {
		areaAfter = *a.AreaAfter
	}

	areaInfo := ""
	if areaBefore != "" || areaAfter != "" {
		areaInfo = fmt.Sprintf("\n면적(㎡): 기존 %s → 변경 후 %s", areaBefore, areaAfter)
	}

	userPrompt := fmt.Sprintf("다음 행정 공고를 쉽게 설명해주세요:\n\n"+
		"구: %s\n사업명: %s\n유형: %s\n조치: %s\n분류: %s\n위치: %s%s",
		a.District, a.Title, a.Type, a.Action, rawCategory, location, areaInfo)

	message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 500,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, &model.ExternalAPIError{Service: "Claude API", Err: err}
	}

	if len(message.Content) == 0 {
		return nil, fmt.Errorf("Claude API 응답이 비어있습니다")
	}

	responseText := message.Content[0].Text
	cleaned := cleanJSON(responseText)

	var result summaryResponse
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		slog.Warn("JSON 파싱 실패, 원본 텍스트 사용", "id", a.ID, "response", responseText)
		result = summaryResponse{
			Summary: responseText,
			Stage:   "분류 불가",
			Related: "",
		}
	}

	return &result, nil
}

// cleanJSON 은 Claude 응답에서 마크다운 코드블록과 추가 텍스트를 제거하고
// 순수 JSON 문자열만 추출한다.
func cleanJSON(s string) string {
	if idx := strings.Index(s, "```"); idx != -1 {
		start := strings.Index(s[idx:], "\n")
		if start != -1 {
			s = s[idx+start+1:]
		}
		if end := strings.Index(s, "```"); end != -1 {
			s = s[:end]
		}
	}

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		s = s[start : end+1]
	}

	return strings.TrimSpace(s)
}

// SummarizeOne 은 단일 공고에 대해 요약을 생성하고 DB에 저장한다.
func (s *SummaryService) SummarizeOne(ctx context.Context, id string) error {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	result, err := s.summarize(ctx, a)
	if err != nil {
		return err
	}

	return s.repo.UpdateSummary(ctx, id, result.Summary, result.Stage, result.Related)
}
