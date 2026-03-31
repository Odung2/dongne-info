package model

import (
	"time"
)

// Announcement 는 서울시 행정 공고 데이터를 나타내는 구조체.
//
// 서울시 열린데이터광장 API에서 수집한 재개발/재건축/도시계획 공고를 저장한다.
// crawler 패키지에서 생성하고, repository로 DB에 저장하며,
// service/api 레이어에서 조회 및 API 응답에 사용한다.
//
// 필드:
//   - District: 서울시 자치구명 ("강남구", "마포구" 등)
//   - Type: 공고 유형. API의 SCLSF(소분류)에서 분류 ("재개발", "재건축", "정비사업")
//   - Action: 조치 유형. API의 RPT_TYPE 원본 ("신설", "변경", "폐지")
//   - Title: 원본 사업명. API의 RGN_NM (예: "은마아파트 재건축 정비구역"). 상세 페이지 하단에 표시.
//   - EasyTitle: AI가 생성한 쉬운 제목 (예: "대치동 은마아파트 재건축 시작돼요"). 카드/알림에 사용.
//   - Location: 위치/주소. API의 PSTN_NM (예: "강남구 대치동 316번지 일대")
//   - Summary: Claude API가 생성한 쉬운 설명 (nil이면 아직 요약되지 않음)
//   - Stage: 재건축 진행 단계 (예: "2/7단계 - 정비구역지정", nil이면 미분류)
//   - Related: 관련 과거 사례. AI가 모르면 빈 문자열 (hallucination 방지)
//   - Impact: "나한테 어떤 의미?" — 이 근처 거주자에게 어떤 영향이 있는지 (예측 아닌 팩트)
//   - ActionTip: "뭘 하면 좋을까?" — 추천 액션 (예: "구청에 문의해보세요")
//   - RawCategory: 원본 API의 대/중/소분류 (예: "의제처리구역 > 정비구역 > 재건축사업구역")
//   - AreaBefore: 기존 면적(㎡). API의 AREA_EXS (예: "179794.9", 신설이면 "0")
//   - AreaAfter: 변경 후 면적(㎡). API의 AREA_CHG_AFTR (예: "181501.5")
//   - SourceURL: 원본 공고 링크 (추후 결정고시 연동 시 사용)
//   - AnnouncedAt: 실제 고시/공고 날짜. API에서 추출. (nil이면 날짜 불명)
//   - SourceID: 중복 방지용 원본 식별자 (예: "rebuild_강남구_11680PPL202501090004")
//   - NotifiedAt: 구독자에게 알림 발송한 시각 (nil이면 아직 미발송)
type Announcement struct {
	ID         string     `db:"id" json:"id"`
	District   string     `db:"district" json:"district"`
	Type       string     `db:"type" json:"type"`
	Action     string     `db:"action" json:"action"`
	Title      string     `db:"title" json:"title"`
	EasyTitle  *string    `db:"easy_title" json:"easy_title,omitempty"`
	Location   *string    `db:"location" json:"location,omitempty"`
	Summary    *string    `db:"summary" json:"summary,omitempty"`
	Stage      *string    `db:"stage" json:"stage,omitempty"`
	Related    *string    `db:"related" json:"related,omitempty"`
	Impact     *string    `db:"impact" json:"impact,omitempty"`
	ActionTip  *string    `db:"action_tip" json:"action_tip,omitempty"`
	RawCategory *string   `db:"raw_category" json:"raw_category,omitempty"`
	AreaBefore  *string   `db:"area_before" json:"area_before,omitempty"`
	AreaAfter   *string   `db:"area_after" json:"area_after,omitempty"`
	AnnouncedAt *time.Time `db:"announced_at" json:"announced_at,omitempty"`
	SourceURL   *string   `db:"source_url" json:"source_url,omitempty"`
	SourceID    *string   `db:"source_id" json:"source_id,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	NotifiedAt *time.Time `db:"notified_at" json:"notified_at,omitempty"`
}

// AnnouncementFilter 는 공고 목록 조회 시 사용하는 필터 조건.
//
// API 핸들러에서 쿼리 파라미터를 파싱하여 이 구조체를 만들고,
// service → repository로 전달하여 조건부 SQL 쿼리를 생성한다.
//
// 사용 예:
//
//	filter := model.AnnouncementFilter{District: "강남구", Type: "재개발", Limit: 20}
//	announcements, err := svc.List(ctx, filter)
//
// 필드:
//   - District: 필터할 자치구명 (빈 문자열이면 전체 조회)
//   - Type: 필터할 공고 유형 (빈 문자열이면 전체 조회)
//   - Limit: 최대 조회 건수 (0 이하면 기본값 20)
//   - Offset: 페이지네이션 오프셋 (0부터 시작)
type AnnouncementFilter struct {
	District string
	Type     string
	Limit    int
	Offset   int
}
