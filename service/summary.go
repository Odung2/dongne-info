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
// 공고 유형(Type)에 따라 다른 프롬프트를 사용하여 최적의 요약을 생성한다.
// - 재개발/재건축: 사업명, 면적, 분류 기반 요약
// - 도시계획: 고시문 전문 기반 요약 (더 풍부한 입력 → 더 좋은 요약)
// - 기타: 범용 프롬프트
//
// 각 요약에는 "나한테 어떤 의미?" (impact)와 "추천 액션" (action_tip)이 포함된다.
//
// 핵심 원칙:
//   - 예측 금지. "집값이 오른다/내린다" 절대 안 함.
//   - 팩트만 요약. 행정 용어를 쉬운 말로 바꾸는 것이 목적.
//   - 모르면 모른다고 한다. 관련 사례를 지어내지 않는다.
type SummaryService struct {
	client *anthropic.Client
	repo   *repository.AnnouncementRepository
}

// NewSummaryService 는 SummaryService를 생성한다.
// apiKey가 빈 문자열이면 nil을 반환한다 (요약 기능 비활성화).
func NewSummaryService(apiKey string, repo *repository.AnnouncementRepository) *SummaryService {
	if apiKey == "" {
		return nil
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &SummaryService{client: &client, repo: repo}
}

// summaryResponse 는 Claude API 응답을 파싱하기 위한 내부 구조체.
// 모든 유형의 프롬프트가 동일한 JSON 구조로 응답하도록 설계.
type summaryResponse struct {
	EasyTitle string `json:"easy_title"` // 쉬운 제목 1줄 (카드/알림용)
	Summary   string `json:"summary"`    // 쉬운 설명 2~3문장
	Stage     string `json:"stage"`      // 진행 단계
	Related   string `json:"related"`    // 관련 사례 (없으면 빈 문자열)
	Impact    string `json:"impact"`     // 나한테 어떤 의미?
	ActionTip string `json:"action_tip"` // 추천 액션
}

// SummarizeUnsummarized 는 아직 AI 요약이 없는 공고들을 찾아 일괄 요약한다.
//
// 공고의 Type에 따라 적절한 프롬프트를 선택하여 요약한다.
// 개별 공고 요약 실패 시에도 나머지는 계속 진행한다.
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

		data := repository.SummaryData{
			EasyTitle: result.EasyTitle,
			Summary:   result.Summary,
			Stage:     result.Stage,
			Related:   result.Related,
			Impact:    result.Impact,
			ActionTip: result.ActionTip,
		}

		if err := s.repo.UpdateSummary(ctx, a.ID, data); err != nil {
			slog.Error("요약 저장 실패", "id", a.ID, "error", err)
			continue
		}

		summarized++
		slog.Info("요약 완료", "id", a.ID, "type", a.Type, "district", a.District, "summary", result.Summary)
	}

	return summarized, nil
}

// summarize 는 공고 유형에 따라 적절한 프롬프트로 Claude API를 호출한다.
func (s *SummaryService) summarize(ctx context.Context, a *model.Announcement) (*summaryResponse, error) {
	var systemPrompt, userPrompt string

	switch a.Type {
	case "재개발", "재건축", "정비사업":
		systemPrompt, userPrompt = s.buildRebuildPrompt(a)
	case "도시계획":
		systemPrompt, userPrompt = s.buildCityplanPrompt(a)
	default:
		systemPrompt, userPrompt = s.buildGeneralPrompt(a)
	}

	message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 700,
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
			Summary:   responseText,
			Stage:     "확인 필요",
			Related:   "",
			Impact:    "",
			ActionTip: "",
		}
	}

	return &result, nil
}

// buildRebuildPrompt 는 재개발/재건축/정비사업 공고용 프롬프트를 생성한다.
//
// 입력: 사업명, 위치, 면적, 분류 (짧은 데이터)
// 7단계: ①기본계획 ②정비구역지정 ③조합설립 ④사업시행인가 ⑤관리처분인가 ⑥착공 ⑦입주
func (s *SummaryService) buildRebuildPrompt(a *model.Announcement) (string, string) {
	system := "서울시 재개발/재건축 공고를 동네 주민이 바로 이해할 수 있게 설명해줘.\n\n" +
		"규칙:\n" +
		"1. 해요체 사용.\n" +
		"2. easy_title: 15자 이내 쉬운 제목. 동네명 + 핵심 변화. 예: \"공덕동 재개발 구역 넓어졌어요\", \"은마아파트 재건축 시작돼요\"\n" +
		"3. summary: 2~3문장. 사업명과 위치를 꼭 포함. 행정 용어는 쉬운 말로.\n" +
		"4. stage: 재개발/재건축 7단계(①기본계획 ②정비구역지정 ③조합설립 ④사업시행인가 ⑤관리처분인가 ⑥착공 ⑦입주) 중 판별 가능하면 \"N/7단계 - 단계명\". 불가능하면 \"확인 필요\".\n" +
		"5. impact: 이 근처 거주자라면 왜 관심 가져야 하는지 1문장. \"~라면 확인해보세요\" 형태. 예측 금지.\n" +
		"6. action_tip: 주민이 할 수 있는 행동 1문장. 구체적으로.\n" +
		"7. related: 같은 지역의 관련 사례를 알면 쓰고, 모르면 빈 문자열. 지어내지 마.\n" +
		"8. 집값 예측 절대 금지.\n\n" +
		"순수 JSON만 응답. 마크다운/추가설명 금지.\n" +
		"{\"easy_title\":\"\",\"summary\":\"\",\"stage\":\"\",\"impact\":\"\",\"action_tip\":\"\",\"related\":\"\"}"

	location := ptrOr(a.Location, "")
	rawCategory := ptrOr(a.RawCategory, "")
	areaBefore := ptrOr(a.AreaBefore, "")
	areaAfter := ptrOr(a.AreaAfter, "")

	areaInfo := ""
	if areaBefore != "" || areaAfter != "" {
		areaInfo = fmt.Sprintf("\n면적(㎡): 기존 %s → 변경 후 %s", areaBefore, areaAfter)
	}

	user := fmt.Sprintf("구: %s\n사업명: %s\n유형: %s\n조치: %s\n분류: %s\n위치: %s%s",
		a.District, a.Title, a.Type, a.Action, rawCategory, location, areaInfo)

	return system, user
}

// buildCityplanPrompt 는 도시계획 결정고시 공고용 프롬프트를 생성한다.
//
// 입력: 고시문 전문 (summary에 임시 저장된 CN 텍스트, 수천 자)
// 5단계: ①입안 ②주민열람 ③위원회심의 ④결정고시 ⑤지형도면고시
func (s *SummaryService) buildCityplanPrompt(a *model.Announcement) (string, string) {
	system := "서울시 도시계획 결정고시를 동네 주민이 바로 이해할 수 있게 설명해줘.\n\n" +
		"규칙:\n" +
		"1. 해요체 사용.\n" +
		"2. easy_title: 15자 이내 쉬운 제목. 동네명 + 핵심 변화. 예: \"동교동 역세권 개발 확정\", \"공덕동 도로 변경돼요\"\n" +
		"3. summary: 2~3문장. 뭘 하겠다는 건지, 어디서, 주민에게 어떤 의미인지 핵심만. 법령 인용/고시번호는 전부 생략.\n" +
		"4. stage: 도시계획 5단계(①입안 ②주민열람/공고 ③위원회심의 ④결정고시 ⑤지형도면고시) 중 판별 가능하면 \"N/5단계 - 단계명\". 불가능하면 \"확인 필요\".\n" +
		"5. impact: 이 근처 거주자/상인이라면 왜 관심 가져야 하는지 1문장. \"~라면 확인해보세요\" 형태.\n" +
		"6. action_tip: 주민이 할 수 있는 행동 1문장. 열람/의견제출 기간이 있으면 꼭 언급.\n" +
		"7. related: 같은 지역의 관련 사례를 알면 쓰고, 모르면 빈 문자열. 지어내지 마.\n" +
		"8. 집값 예측 절대 금지. 법률 조항 인용 금지. 고시번호 언급 금지.\n\n" +
		"순수 JSON만 응답. 마크다운/추가설명 금지.\n" +
		"{\"easy_title\":\"\",\"summary\":\"\",\"stage\":\"\",\"impact\":\"\",\"action_tip\":\"\",\"related\":\"\"}"

	// 도시계획은 summary에 고시문 전문이 임시 저장되어 있음
	content := ptrOr(a.Summary, "")
	if len(content) > 3000 {
		content = content[:3000] + "..."
	}

	user := fmt.Sprintf("구: %s\n제목: %s\n고시유형: %s\n\n[고시문 전문]\n%s",
		a.District, a.Title, a.Action, content)

	return system, user
}

// buildGeneralPrompt 는 기타 유형의 공고용 범용 프롬프트를 생성한다.
// 추후 교통, 교육 등 새 장르 추가 시 장르별 프롬프트를 만들기 전 임시로 사용.
func (s *SummaryService) buildGeneralPrompt(a *model.Announcement) (string, string) {
	system := "서울시 행정 공고를 동네 주민이 바로 이해할 수 있게 설명해줘.\n\n" +
		"규칙:\n" +
		"1. 해요체 사용.\n" +
		"2. easy_title: 15자 이내 쉬운 제목. 동네명 + 핵심 변화.\n" +
		"3. summary: 2~3문장 핵심 요약. 행정 용어는 쉬운 말로.\n" +
		"4. stage: 진행 단계를 알 수 있으면 표시, 모르면 \"확인 필요\".\n" +
		"5. impact: 주민에게 어떤 의미인지 1문장. \"~라면 확인해보세요\" 형태.\n" +
		"6. action_tip: 주민이 할 수 있는 행동 1문장.\n" +
		"7. related: 관련 사례. 모르면 빈 문자열.\n" +
		"8. 집값 예측 절대 금지.\n\n" +
		"순수 JSON만 응답.\n" +
		"{\"easy_title\":\"\",\"summary\":\"\",\"stage\":\"\",\"impact\":\"\",\"action_tip\":\"\",\"related\":\"\"}"

	user := fmt.Sprintf("구: %s\n제목: %s\n유형: %s\n조치: %s",
		a.District, a.Title, a.Type, a.Action)

	return system, user
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

// ptrOr 은 포인터가 nil이면 기본값을, 아니면 값을 반환한다.
func ptrOr(p *string, defaultVal string) string {
	if p == nil {
		return defaultVal
	}
	return *p
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

	return s.repo.UpdateSummary(ctx, id, repository.SummaryData{
		EasyTitle: result.EasyTitle,
		Summary:   result.Summary,
		Stage:     result.Stage,
		Related:   result.Related,
		Impact:    result.Impact,
		ActionTip: result.ActionTip,
	})
}
