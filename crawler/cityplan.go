package crawler

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"dongne-info/model"
	"dongne-info/repository"
)

// cityplanRow 는 서울시 upisAnnouncement API 응답의 개별 행을 나타내는 구조체.
//
// 도시계획 결정고시 데이터를 파싱하여 model.Announcement로 변환한다.
// upisRebuild와 달리 고시문 전문(CN)이 포함되어 있어 AI 요약의 입력이 풍부하다.
//
// 필드:
//   - AncmntMngCD: 고시관리코드 → SourceID 생성에 사용 (중복 방지)
//   - PrjcCD: 프로젝트코드 (참고용)
//   - AncmntType: 고시유형 ("결정+지적", "실시" 등) → Action
//   - AncmntNo: 고시번호 (예: "2026-152")
//   - AncmntYMD: 고시일자 (예: "2026-03-18T00:00:00.000")
//   - AncmntInst: 고시기관 ("서울특별시", "강남구" 등) → 구 필터링에 사용
//   - TTL: 제목 (예: "도시관리계획 결정(변경) 및 지형도면 고시") → Title
//   - CN: 내용 (고시문 전문) → AI 요약의 핵심 입력. 최대 수천 자.
type cityplanRow struct {
	AncmntMngCD string `xml:"ANCMNT_MNG_CD"`
	PrjcCD      string `xml:"PRJC_CD"`
	AncmntType  string `xml:"ANCMNT_TYPE"`
	AncmntNo    string `xml:"ANCMNT_NO"`
	AncmntYMD   string `xml:"ANCMNT_YMD"`
	AncmntInst  string `xml:"ANCMNT_INST"`
	TTL         string `xml:"TTL"`
	CN          string `xml:"CN"`
}

// cityplanResult 는 서울시 upisAnnouncement API 응답 전체를 나타내는 구조체.
type cityplanResult struct {
	TotalCount int            `xml:"list_total_count"`
	Rows       []cityplanRow  `xml:"row"`
}

// CityplanCrawler 는 서울시 도시계획 결정고시 데이터를 수집하는 크롤러.
//
// 서울시 열린데이터광장의 upisAnnouncement API를 호출하여
// 도시계획 결정고시(용도지역 변경, 지구단위계획, 도시계획시설 등)를 수집한다.
//
// upisRebuild와의 차이:
//   - 구 필터 파라미터가 없어서 전체 데이터를 가져온 후 ANCMNT_INST로 필터링
//   - 고시문 전문(CN)이 포함되어 AI 요약 품질이 높음
//   - 최신 데이터부터 가져오기 위해 뒤에서부터 페이징
//
// 사용 예:
//
//	crawler := crawler.NewCityplanCrawler(apiKey, repo)
//	count, err := crawler.Crawl(ctx, []string{"강남구", "마포구"})
type CityplanCrawler struct {
	apiKey string
	repo   *repository.AnnouncementRepository
}

// NewCityplanCrawler 는 CityplanCrawler를 생성한다.
func NewCityplanCrawler(apiKey string, repo *repository.AnnouncementRepository) *CityplanCrawler {
	return &CityplanCrawler{apiKey: apiKey, repo: repo}
}

// Crawl 는 도시계획 결정고시를 수집하고 대상 구에 해당하는 공고만 저장한다.
//
// API에 구 필터가 없으므로, 최신 100건을 가져온 후
// ANCMNT_INST(고시기관)에 대상 구가 포함된 것만 필터링하여 저장한다.
//
// 파라미터:
//   - districts: 수집 대상 자치구 목록 (예: ["강남구", "마포구"])
//
// 반환:
//   - int: 새로 저장된 공고 수
//   - error: API 호출 자체가 실패한 경우에만 에러 반환
func (c *CityplanCrawler) Crawl(ctx context.Context, districts []string) (int, error) {
	// 앞쪽이 최신 데이터. 최신 500건을 가져와서 대상 구 필터링.
	// (도시계획 고시는 전체 구 통합이라 강남/마포 비율이 낮아 넉넉하게 가져옴)
	rows, err := c.fetchFromAPI(1, 500)
	if err != nil {
		return 0, &model.ExternalAPIError{Service: "서울시 upisAnnouncement API", Err: err}
	}

	slog.Info("도시계획 API 응답", "fetched", len(rows))

	// 대상 구 필터링을 위한 맵
	districtSet := make(map[string]bool)
	for _, d := range districts {
		districtSet[d] = true
	}

	newCount := 0
	for _, row := range rows {
		// 고시기관으로 구 필터링
		district := matchDistrict(row.AncmntInst, districtSet)
		if district == "" {
			continue
		}

		sourceID := fmt.Sprintf("cityplan_%s", row.AncmntMngCD)

		exists, err := c.repo.ExistsBySourceID(ctx, sourceID)
		if err != nil {
			slog.Error("중복 확인 실패", "source_id", sourceID, "error", err)
			continue
		}
		if exists {
			continue
		}

		// CN(내용)이 너무 길면 앞부분만 사용 (AI 토큰 절약)
		content := row.CN
		if len(content) > 2000 {
			content = content[:2000] + "..."
		}

		// 고시일자 파싱 (형식: "2026-03-18T00:00:00.000")
		var announcedAt *time.Time
		if len(row.AncmntYMD) >= 10 {
			if t, err := time.Parse("2006-01-02", row.AncmntYMD[:10]); err == nil {
				announcedAt = &t
			}
		}

		announcement := &model.Announcement{
			District:    district,
			Type:        "도시계획",
			Action:      row.AncmntType,
			Title:       row.TTL,
			Summary:     strPtr(content),
			AnnouncedAt: announcedAt,
			RawCategory: strPtr(fmt.Sprintf("고시번호: %s | 고시일: %s", row.AncmntNo, row.AncmntYMD[:10])),
			SourceID:    strPtr(sourceID),
		}

		if err := c.repo.Create(ctx, announcement); err != nil {
			slog.Error("도시계획 공고 저장 실패", "source_id", sourceID, "error", err)
			continue
		}

		newCount++
		slog.Info("도시계획 공고 저장", "district", district, "title", row.TTL, "no", row.AncmntNo)
	}

	return newCount, nil
}

// fetchTotalCount 는 API의 총 데이터 건수를 조회한다.
//
// 1건만 요청하여 list_total_count를 확인하고,
// 최신 데이터부터 가져오기 위한 시작 인덱스를 계산하는 데 사용한다.
func (c *CityplanCrawler) fetchTotalCount() (int, error) {
	url := fmt.Sprintf("%s/%s/xml/upisAnnouncement/1/1/", seoulAPIBaseURL, c.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var result cityplanResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	return result.TotalCount, nil
}

// fetchFromAPI 는 지정된 범위의 도시계획 고시 데이터를 조회한다.
func (c *CityplanCrawler) fetchFromAPI(startIdx, endIdx int) ([]cityplanRow, error) {
	url := fmt.Sprintf("%s/%s/xml/upisAnnouncement/%d/%d/",
		seoulAPIBaseURL, c.apiKey, startIdx, endIdx)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var result cityplanResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	return result.Rows, nil
}

// matchDistrict 는 고시기관명에서 대상 자치구를 찾아 반환한다.
//
// 고시기관이 "서울특별시"인 경우 시 전체 공고이므로 빈 문자열을 반환한다.
// "강남구", "강남구청" 등이 포함되면 해당 구명을 반환한다.
//
// 예:
//   - "강남구" → "강남구"
//   - "서울특별시" → "" (시 전체 공고, 필터링 제외)
//   - "마포구청" → "마포구"
func matchDistrict(inst string, districtSet map[string]bool) string {
	for district := range districtSet {
		if strings.Contains(inst, district) {
			return district
		}
	}
	return ""
}
