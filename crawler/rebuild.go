// Package crawler 는 서울시 공공 API에서 행정 공고 데이터를 수집한다.
//
// 매일 자동 실행되어 새 공고를 감지하고 DB에 저장하는 역할을 한다.
// 현재 지원하는 데이터 소스:
//   - upisRebuild: 재개발/재건축 정비사업 (이 파일)
//   - upisCityplan: 도시계획 결정고시 (cityplan.go, 추가 예정)
//   - 자치구 공고: 구청 홈페이지 크롤링 (district.go, 추가 예정)
//
// 크롤링 순서:
//  1. 서울시 API 호출 (XML 응답)
//  2. source_id로 DB에 이미 있는지 중복 체크
//  3. 새 공고만 DB에 저장
//  4. (별도 프로세스에서) Claude API로 AI 요약 생성
//  5. (별도 프로세스에서) 구독자에게 알림 발송
package crawler

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"dongne-info/model"
	"dongne-info/repository"
)

// seoulAPIBaseURL 은 서울시 열린데이터광장 API의 기본 URL.
const seoulAPIBaseURL = "http://openapi.seoul.go.kr:8088"

// xmlRow 는 서울시 upisRebuild API 응답의 개별 행을 나타내는 구조체.
//
// API 응답 XML의 <row> 요소와 매핑된다.
// 이 구조체는 API 응답 파싱 전용이며, 파싱 후 model.Announcement로 변환하여 DB에 저장한다.
//
// 필드:
//   - RptType: 조치 유형 ("신설", "변경", "폐지") → Announcement.Action
//   - PrjcCD: 사업 코드 → SourceID 생성에 사용 (중복 방지)
//   - Lclsf: 대분류 ("의제처리구역" 등) → RawCategory에 포함
//   - Mclsf: 중분류 ("정비구역" 등) → RawCategory에 포함
//   - Sclsf: 소분류 ("재건축사업구역" 등) → Type 분류 + RawCategory에 포함
//   - PstnNm: 위치/주소 (예: "강남구 대치동 316번지 일대") → Location
//   - RgnNm: 사업명 (예: "은마아파트 재건축 정비구역") → Title ← AI 요약의 핵심 입력
//   - AreaExs: 기존 면적(㎡) → AreaBefore (신설이면 "0")
//   - AreaChg: 면적 변동량(㎡) (사용 안 함, before/after로 계산 가능)
//   - AreaChgAftr: 변경 후 면적(㎡) → AreaAfter
//   - DcsnCode: 결정고시 관리코드 → SourceURL에 저장 (추후 원문 링크 연결용)
//   - AreaExs: 기존 면적
//   - AreaChg: 면적 변동량
//   - AreaChgAftr: 변경 후 면적
type xmlRow struct {
	RptType     string `xml:"RPT_TYPE"`
	PrjcCD      string `xml:"PRJC_CD"`
	Lclsf       string `xml:"LCLSF"`
	Mclsf       string `xml:"MCLSF"`
	Sclsf       string `xml:"SCLSF"`
	PstnNm      string `xml:"PSTN_NM"`
	RgnNm       string `xml:"RGN_NM"`
	AreaExs     string `xml:"AREA_EXS"`
	AreaChg     string `xml:"AREA_CHG"`
	AreaChgAftr string `xml:"AREA_CHG_AFTR"`
	DcsnCode    string `xml:"DCSN_ANCMNT_MNG_CD"`
}

// xmlResult 는 서울시 upisRebuild API 응답 전체를 나타내는 구조체.
//
// API 응답 XML의 루트 요소와 매핑된다.
type xmlResult struct {
	TotalCount int      `xml:"list_total_count"`
	Rows       []xmlRow `xml:"row"`
}

// RebuildCrawler 는 서울시 재개발/재건축 데이터를 수집하는 크롤러.
//
// 서울시 열린데이터광장의 upisRebuild API를 호출하여
// 강남구/마포구의 재개발·재건축 정비사업 데이터를 수집하고 DB에 저장한다.
//
// cmd/crawl/main.go에서 생성하여 사용한다.
//
// 사용 예:
//
//	repo := repository.NewAnnouncementRepository(db)
//	crawler := crawler.NewRebuildCrawler(apiKey, repo)
//	count, err := crawler.Crawl(ctx, "강남구")
type RebuildCrawler struct {
	apiKey string
	repo   *repository.AnnouncementRepository
}

// NewRebuildCrawler 는 RebuildCrawler를 생성한다.
//
// 파라미터:
//   - apiKey: 서울시 열린데이터광장 API 인증키 (SEOUL_API_KEY 환경변수)
//   - repo: 공고 저장에 사용할 AnnouncementRepository
func NewRebuildCrawler(apiKey string, repo *repository.AnnouncementRepository) *RebuildCrawler {
	return &RebuildCrawler{apiKey: apiKey, repo: repo}
}

// Crawl 는 지정된 자치구의 재개발 데이터를 수집하고 새로 저장된 공고 수를 반환한다.
//
// 실행 흐름:
//  1. 서울시 API에서 해당 구의 재개발 데이터를 XML으로 조회
//  2. 각 행의 사업 코드(PrjcCD)로 source_id를 생성
//  3. source_id가 DB에 이미 있으면 건너뜀 (중복 방지)
//  4. 새 공고만 model.Announcement로 변환하여 DB에 저장
//
// 에러 발생 시에도 개별 행 단위로 처리하여, 일부 실패해도 나머지는 계속 저장한다.
//
// 파라미터:
//   - district: 크롤링할 자치구명 (예: "강남구", "마포구")
//
// 반환:
//   - int: 새로 저장된 공고 수
//   - error: API 호출 자체가 실패한 경우에만 에러 반환
func (c *RebuildCrawler) Crawl(ctx context.Context, district string) (int, error) {
	rows, err := c.fetchFromAPI(district)
	if err != nil {
		return 0, &model.ExternalAPIError{Service: "서울시 upisRebuild API", Err: err}
	}

	newCount := 0
	for _, row := range rows {
		sourceID := fmt.Sprintf("rebuild_%s_%s", district, row.PrjcCD)

		exists, err := c.repo.ExistsBySourceID(ctx, sourceID)
		if err != nil {
			slog.Error("중복 확인 실패", "source_id", sourceID, "error", err)
			continue
		}
		if exists {
			continue
		}

		// RgnNm(사업명)이 있으면 제목으로 사용, 없으면 분류 조합
		title := row.RgnNm
		if title == "" {
			title = fmt.Sprintf("%s %s %s", row.Mclsf, row.Sclsf, row.RptType)
		}

		// 결정고시 코드가 있으면 정비사업 정보몽땅 검색 URL로 저장
		// 추후 원문 직접 링크가 확인되면 교체 예정
		var sourceURL *string
		if row.DcsnCode != "" {
			url := fmt.Sprintf("https://cleanup.seoul.go.kr/search/searchList.do?searchText=%s", row.DcsnCode)
			sourceURL = &url
		}

		announcement := &model.Announcement{
			District:    district,
			Type:        classifyType(row.Sclsf),
			Action:      row.RptType,
			Title:       title,
			Location:    strPtr(row.PstnNm),
			SourceURL:   sourceURL,
			SourceID:    strPtr(sourceID),
			RawCategory: strPtr(fmt.Sprintf("%s > %s > %s", row.Lclsf, row.Mclsf, row.Sclsf)),
			AreaBefore:  strPtr(row.AreaExs),
			AreaAfter:   strPtr(row.AreaChgAftr),
		}

		if err := c.repo.Create(ctx, announcement); err != nil {
			slog.Error("공고 저장 실패", "source_id", sourceID, "error", err)
			continue
		}

		newCount++
		slog.Info("새 공고 저장", "district", district, "location", row.PstnNm, "type", row.RptType)
	}

	return newCount, nil
}

// fetchFromAPI 는 서울시 upisRebuild API를 호출하여 원시 데이터를 반환한다.
//
// 최대 100건을 조회하며, XML 응답을 xmlResult 구조체로 파싱한다.
func (c *RebuildCrawler) fetchFromAPI(district string) ([]xmlRow, error) {
	url := fmt.Sprintf("%s/%s/xml/upisRebuild/1/100/%s/",
		seoulAPIBaseURL, c.apiKey, district)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var result xmlResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	slog.Info("API 응답", "district", district, "total_count", result.TotalCount)
	return result.Rows, nil
}

// classifyType 은 서울시 API의 SCLSF(소분류) 필드를 분석해서
// "재개발", "재건축" 등으로 분류한다.
//
// 소분류 예시:
//   - "재건축사업구역", "주택재건축사업" → "재건축"
//   - "재개발사업구역", "주택재개발사업" → "재개발"
//   - 그 외 → "정비사업"
// classifyType 은 서울시 API의 SCLSF(소분류) 필드를 분석해서
// "재개발", "재건축", "도심 재개발" 등으로 분류한다.
//
// 소분류 예시:
//   - "재건축사업구역", "주택재건축사업" → "재건축"
//   - "주택재개발사업구역" → "재개발"
//   - "도시환경정비사업구역" → "도심 재개발" (상업지역 재개발과 유사)
//   - "주거환경개선사업구역" → "주거환경개선"
func classifyType(sclsf string) string {
	switch {
	case contains(sclsf, "재건축"):
		return "재건축"
	case contains(sclsf, "재개발"):
		return "재개발"
	case contains(sclsf, "도시환경정비"):
		return "도심 재개발"
	case contains(sclsf, "주거환경개선"):
		return "주거환경개선"
	default:
		return "재개발"
	}
}

// classifyAction 은 서울시 API의 RPT_TYPE(보고 유형) 필드를 분석해서
// "신설", "변경", "폐지", "기타" 중 하나로 분류한다.
//
// RPT_TYPE에 포함된 키워드로 판단한다:
//   - "신설", "지정", "설립" → "신설"
//   - "변경", "수정" → "변경"
//   - "폐지", "해제" → "폐지"
//   - 그 외 → "기타"
func classifyAction(rptType string) string {
	switch {
	case contains(rptType, "신설", "지정", "설립"):
		return "신설"
	case contains(rptType, "변경", "수정"):
		return "변경"
	case contains(rptType, "폐지", "해제"):
		return "폐지"
	default:
		return "기타"
	}
}

// contains 는 문자열 s에 substrs 중 하나라도 포함되어 있는지 확인한다.
func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		for i := range s {
			if i+len(sub) <= len(s) && s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}

// strPtr 는 빈 문자열이면 nil, 아니면 문자열 포인터를 반환한다.
//
// DB에 빈 문자열 대신 NULL을 저장하기 위해 사용한다.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
