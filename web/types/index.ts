/** Go의 model.Announcement와 1:1 매핑 */
export interface Announcement {
  id: string;
  district: string;        // "강남구", "마포구"
  type: string;            // "재개발", "재건축", "도시계획"
  action: string;          // "신설", "변경", "폐지"
  title: string;           // 원본 사업명 (API의 RGN_NM)
  easy_title?: string;     // AI가 만든 쉬운 제목 (카드/알림용)
  location?: string;       // 위치 (API의 PSTN_NM)
  summary?: string;        // AI 쉬운 설명
  stage?: string;          // 진행 단계 (예: "2/7단계 - 정비구역지정")
  related?: string;        // 관련 사례
  impact?: string;         // "나한테 어떤 의미?"
  action_tip?: string;     // "추천 액션"
  raw_category?: string;   // 원본 분류 (대>중>소)
  area_before?: string;    // 기존 면적(㎡)
  area_after?: string;     // 변경 후 면적(㎡)
  announced_at?: string;   // 실제 고시/공고 날짜
  source_url?: string;     // 원문 링크
  created_at: string;      // DB 저장 시각
}

/** 반응 집계 */
export interface ReactionSummary {
  interest: number;
  agree: number;
  disagree: number;
}

/** 구 정보 — URL 슬러그와 한글명 매핑 */
export interface DistrictInfo {
  slug: string;    // "gangnam", "mapo"
  name: string;    // "강남구", "마포구"
}

/** 지원하는 구 목록 */
export const DISTRICTS: DistrictInfo[] = [
  { slug: "gangnam", name: "강남구" },
  { slug: "mapo", name: "마포구" },
];

/** 슬러그 → 한글명 변환. 없으면 null */
export function getDistrictName(slug: string): string | null {
  return DISTRICTS.find((d) => d.slug === slug)?.name ?? null;
}

/** 한글명 → 슬러그 변환. 없으면 null */
export function getDistrictSlug(name: string): string | null {
  return DISTRICTS.find((d) => d.name === name)?.slug ?? null;
}
