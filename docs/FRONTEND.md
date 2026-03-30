# 프론트엔드 개발 계획 (Frontend Development)

> Next.js 14+ App Router / Tailwind CSS / 모바일 퍼스트

---

## 기술 스택

| 영역 | 기술 | 이유 |
|---|---|---|
| 프레임워크 | Next.js 14+ (App Router) | SSR로 SEO 확보, 구글 검색 유입 핵심 |
| 스타일 | Tailwind CSS | 빠른 개발, 모바일 우선 설계 |
| API 통신 | Next.js API Routes (프록시) | CORS 문제 없음, Go API URL 비노출 |
| 배포 | Vercel | Next.js 최적화, 무료 플랜 |
| 상태관리 | 없음 (서버 컴포넌트 위주) | 이 규모에서 상태관리 라이브러리 불필요 |

---

## 디렉토리 구조

```
dongne-info/
└── web/                          ← Next.js 프로젝트 루트
    ├── app/
    │   ├── layout.tsx            ← 전체 레이아웃 (헤더, 푸터)
    │   ├── page.tsx              ← 랜딩 페이지 (/)
    │   ├── gangnam/
    │   │   ├── page.tsx          ← 강남구 공고 목록 (/gangnam)
    │   │   └── [id]/
    │   │       └── page.tsx      ← 공고 상세 (/gangnam/[id])
    │   ├── mapo/
    │   │   ├── page.tsx          ← 마포구 공고 목록 (/mapo)
    │   │   └── [id]/
    │   │       └── page.tsx      ← 공고 상세 (/mapo/[id])
    │   └── api/
    │       ├── announcements/
    │       │   └── route.ts      ← GET 프록시 → Go API
    │       ├── announcements/[id]/
    │       │   └── route.ts      ← GET 프록시 → Go API
    │       ├── announcements/[id]/react/
    │       │   └── route.ts      ← POST 프록시 → Go API
    │       ├── announcements/[id]/reactions/
    │       │   └── route.ts      ← GET 프록시 → Go API
    │       └── subscribe/
    │           └── route.ts      ← POST 프록시 → Go API
    ├── components/
    │   ├── layout/
    │   │   ├── Header.tsx        ← 상단 네비게이션
    │   │   └── Footer.tsx        ← 하단 (서비스 소개, 링크)
    │   ├── announcement/
    │   │   ├── AnnouncementCard.tsx   ← 공고 카드 (목록에서 사용)
    │   │   ├── AnnouncementDetail.tsx ← 공고 상세 본문
    │   │   ├── StageBar.tsx          ← 7단계 프로그레스 바
    │   │   ├── AreaChange.tsx        ← 면적 변동 표시
    │   │   └── TypeBadge.tsx         ← 유형/조치 뱃지
    │   ├── reaction/
    │   │   ├── InterestButton.tsx    ← [📌 관심 있어요] 버튼
    │   │   └── ReactionSummary.tsx   ← 반응 집계 표시
    │   ├── subscribe/
    │   │   └── SubscribeForm.tsx     ← 인라인 구독 폼
    │   └── share/
    │       └── KakaoShare.tsx        ← 카카오톡 공유 버튼
    ├── lib/
    │   └── api.ts                ← Go API 호출 유틸 (서버/클라이언트 공용)
    ├── types/
    │   └── index.ts              ← TypeScript 타입 정의
    ├── tailwind.config.ts
    ├── next.config.ts
    ├── package.json
    └── tsconfig.json
```

---

## 페이지별 상세 설계

### 1. 랜딩 페이지 (`/`)

```
┌─────────────────────────────────┐
│         동네정보                  │
│  우리 동네 행정 공고, 쉽게 알려드려요  │
│                                 │
│  ┌───────────┐ ┌───────────┐   │
│  │  강남구 →  │ │  마포구 →  │   │
│  │  공고 13건  │ │  공고 47건  │   │
│  └───────────┘ └───────────┘   │
│                                 │
│  📌 최근 인기 공고                │
│  ┌─────────────────────────┐   │
│  │ 은마아파트 재건축 정비구역    │   │
│  │ 관심 412명                │   │
│  └─────────────────────────┘   │
│  ┌─────────────────────────┐   │
│  │ 개포주공4단지 재건축        │   │
│  │ 관심 231명                │   │
│  └─────────────────────────┘   │
└─────────────────────────────────┘
```

- 서버 컴포넌트 (SSR, SEO)
- Go API에서 각 구의 공고 수 + 최근 인기 공고 3건 조회
- 구 카드 클릭 → `/gangnam` 또는 `/mapo`로 이동

### 2. 공고 목록 페이지 (`/gangnam`, `/mapo`)

```
┌─────────────────────────────────┐
│  ← 강남구 공고                    │
│                                 │
│  [전체] [재개발] [재건축] [도시계획] │  ← 필터 탭
│                                 │
│  ┌─────────────────────────┐   │
│  │ 재건축 │ 신설              │   │  ← 유형 뱃지 + 조치 뱃지
│  │                          │   │
│  │ 은마아파트 재건축 정비구역    │   │  ← 사업명 (Title)
│  │ 강남구 대치동 316번지 일대   │   │  ← 위치
│  │                          │   │
│  │ 강남구 대치동 은마아파트 일대  │   │  ← AI 요약 미리보기 (2줄)
│  │ 를 재건축할 수 있게 공식으...  │   │
│  │                          │   │
│  │ 📊 2/7단계  📌 412명      │   │  ← 단계 + 관심 수
│  └─────────────────────────┘   │
│                                 │
│  ┌─────────────────────────┐   │
│  │ 재건축 │ 변경              │   │
│  │ 개포주공4단지 주택재건축...   │   │
│  │ ...                       │   │
│  └─────────────────────────┘   │
│                                 │
│  ┌─────────────────────────┐   │
│  │ 📬 새 공고 알림 받기        │   │  ← 인라인 구독 폼
│  │ 이메일: [          ] [구독] │   │
│  └─────────────────────────┘   │
└─────────────────────────────────┘
```

- 서버 컴포넌트 (SSR, SEO)
- 필터 탭: 클라이언트 컴포넌트 (URL 쿼리 파라미터로 필터)
- 카드 클릭 → `/gangnam/[id]`로 이동
- 하단 구독 폼: 클라이언트 컴포넌트

#### 컴포넌트 분해

```
page.tsx (서버)
├── FilterTabs (클라이언트) — 유형 필터 탭
├── AnnouncementCard × N (서버) — 공고 카드 목록
│   ├── TypeBadge — 유형 뱃지 (재개발/재건축/도시계획)
│   ├── TypeBadge — 조치 뱃지 (신설/변경/폐지)
│   └── ReactionSummary — 관심 수 (서버에서 미리 조회)
└── SubscribeForm (클라이언트) — 구독 폼
```

### 3. 공고 상세 페이지 (`/gangnam/[id]`, `/mapo/[id]`)

```
┌─────────────────────────────────┐
│  ← 강남구                        │
│                                 │
│  재건축 │ 신설                    │  ← 뱃지
│                                 │
│  은마아파트 재건축 정비구역         │  ← 사업명
│  강남구 대치동 316번지 일대        │  ← 위치
│                                 │
│  ─────────────────────────────  │
│                                 │
│  💬 쉬운 설명                     │
│  강남구 대치동 은마아파트 일대를     │
│  재건축할 수 있게 공식으로 정했어요.  │
│  약 24만 3천㎡ 규모로 낡은 건물을   │
│  새로 지을 준비가 시작된 거예요.    │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  📊 진행 단계                     │
│  ■■□□□□□ 2/7단계               │  ← 단계 있을 때만 프로그레스 바
│  정비구역 지정                    │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  📐 면적 변동                     │  ← 면적 있을 때만 표시
│  기존: 0㎡ → 변경 후: 243,552㎡   │
│  (신규 지정)                     │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  ┌─────────────────────────┐   │
│  │  📌 관심 있어요  412명     │   │  ← 관심 버튼 (클릭하면 +1)
│  └─────────────────────────┘   │
│                                 │
│  ┌──────────┐ ┌──────────┐    │
│  │ 📤 공유   │ │ 📄 원문   │    │  ← 카카오 공유 + 원문 링크
│  └──────────┘ └──────────┘    │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  분류: 의제처리구역 > 정비구역 >   │  ← 원본 분류 정보
│       재건축사업구역               │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  📬 이 공고 관련 소식 받기         │  ← 구독 유도
│  [이메일 입력] [알림 받기]         │
└─────────────────────────────────┘
```

- 서버 컴포넌트 (SSR, SEO — 공고 상세는 구글 검색 유입의 핵심)
- 관심 버튼: 클라이언트 컴포넌트 (session_id 기반, 비로그인 가능)
- 카카오 공유: 클라이언트 컴포넌트

#### 컴포넌트 분해

```
page.tsx (서버)
├── TypeBadge × 2 — 유형 + 조치 뱃지
├── AnnouncementDetail (서버) — 요약/위치/분류
├── StageBar (서버) — 진행 단계 (있을 때만)
├── AreaChange (서버) — 면적 변동 (있을 때만)
├── InterestButton (클라이언트) — 관심 표시 버튼
├── ReactionSummary (서버) — 반응 집계
├── KakaoShare (클라이언트) — 카카오 공유
└── SubscribeForm (클라이언트) — 구독 폼
```

#### 조건부 표시 규칙

```
StageBar:
  - stage가 "N/7단계"로 시작하면 → 프로그레스 바 + 텍스트
  - stage가 "확인 필요"면 → 텍스트만 ("진행 단계 확인 필요")
  - stage가 null이면 → 섹션 자체를 숨김

AreaChange:
  - area_before와 area_after 둘 다 있으면 → 변동 표시
  - area_before가 "0"이면 → "신규 지정 (243,552㎡)"
  - 둘 다 없으면 → 섹션 자체를 숨김

Related:
  - 빈 문자열이면 → 표시 안 함
  - 값 있으면 → "📎 관련 사례" 섹션에 표시
```

---

## API 프록시 설계

Next.js API Routes가 Go 서버를 프록시하는 구조.
프론트에서는 `/api/...`만 호출하고, Go 서버 URL은 노출하지 않는다.

```
프론트 (브라우저)                  Next.js 서버                Go API 서버
─────────────                  ──────────                ──────────
GET /api/announcements    →    API Route에서              →  GET localhost:8080/api/announcements
                               Go API 호출 후 응답 전달

POST /api/subscribe       →    API Route에서              →  POST localhost:8080/api/subscribe
                               Go API 호출 후 응답 전달
```

```typescript
// web/lib/api.ts

const API_BASE = process.env.GO_API_URL || "http://localhost:8080";

export async function fetchAnnouncements(district: string, type?: string) {
  const params = new URLSearchParams({ district });
  if (type) params.set("type", type);
  const res = await fetch(`${API_BASE}/api/announcements?${params}`);
  return res.json();
}
```

환경변수:
```
# web/.env.local (개발)
GO_API_URL=http://localhost:8080

# Vercel (배포)
GO_API_URL=https://dongne-info-api.railway.app
```

---

## TypeScript 타입 정의

```typescript
// web/types/index.ts

// Go의 model.Announcement와 1:1 매핑
export interface Announcement {
  id: string;
  district: string;          // "강남구", "마포구"
  type: string;              // "재개발", "재건축", "도시계획"
  action: string;            // "신설", "변경", "폐지"
  title: string;             // 사업명 (RGN_NM)
  location?: string;         // 위치
  summary?: string;          // AI 쉬운 설명
  stage?: string;            // 진행 단계
  related?: string;          // 관련 사례
  raw_category?: string;     // 원본 분류
  area_before?: string;      // 기존 면적(㎡)
  area_after?: string;       // 변경 후 면적(㎡)
  source_url?: string;       // 원문 링크
  created_at: string;        // ISO 8601
}

export interface ReactionSummary {
  interest: number;
  agree: number;
  disagree: number;
}

export interface CreateSubscriberRequest {
  contact: string;
  type: "email" | "kakao";
  districts: string[];
}
```

---

## 디자인 원칙

### 모바일 퍼스트

```
타겟 유저: 30~50대, 카카오톡에서 링크 클릭해서 들어옴
→ 모바일 화면이 메인. 데스크탑은 넓어지기만 하면 됨.
→ Tailwind의 sm:/md:/lg: 브레이크포인트 활용
→ 카드, 버튼 터치 영역 최소 44px
```

### 톤앤매너

```
- 깔끔하고 신뢰감 있는 디자인 (공익 서비스)
- 화려하지 않게. 정보가 주인공.
- 색상: 메인 블루(신뢰) + 뱃지별 컬러
  - 재개발: 파란색
  - 재건축: 초록색
  - 도시계획: 보라색
  - 신설: 초록 뱃지
  - 변경: 노란 뱃지
  - 폐지: 빨간 뱃지
```

### SEO 최적화

```
- 서버 컴포넌트로 HTML 렌더링 (검색 엔진이 읽을 수 있도록)
- 메타 태그: 각 페이지별 title, description, og:image
- 구조화 데이터 (JSON-LD): 공고 상세 페이지
- URL 구조: /gangnam/[id] — 깔끔하고 의미 있는 URL

예시 메타:
  title: "은마아파트 재건축 정비구역 | 강남구 공고 - 동네정보"
  description: "강남구 대치동 은마아파트 일대를 재건축할 수 있게 공식으로 정했어요..."
```

---

## 구현 순서

### 1단계: 프로젝트 세팅 + 레이아웃 (Day 1)

```
- Next.js 프로젝트 생성 (web/)
- Tailwind CSS 설정
- 전체 레이아웃 (Header, Footer)
- API 프록시 유틸 (lib/api.ts)
- TypeScript 타입 정의
- 환경변수 설정 (.env.local)
```

### 2단계: 공고 목록 페이지 (Day 2-3)

```
- 랜딩 페이지 (구 선택 + 인기 공고)
- 공고 목록 페이지 (/gangnam, /mapo)
- AnnouncementCard 컴포넌트
- TypeBadge 컴포넌트
- 유형 필터 탭
- 모바일 반응형 확인
```

### 3단계: 공고 상세 페이지 (Day 4-5)

```
- AnnouncementDetail 컴포넌트
- StageBar 컴포넌트 (조건부 표시)
- AreaChange 컴포넌트 (조건부 표시)
- SEO 메타 태그
- 모바일 반응형 확인
```

### 4단계: 관심 표시 + 반응 (Day 6)

```
- InterestButton 클라이언트 컴포넌트
- session_id 생성 (쿠키 기반)
- 클릭 시 POST /api/announcements/:id/react
- 반응 집계 실시간 업데이트
- 이미 눌렀으면 비활성화
```

### 5단계: 구독 폼 (Day 7)

```
- SubscribeForm 클라이언트 컴포넌트
- 이메일 입력 + 구 자동 선택 (현재 보고 있는 구)
- POST /api/subscribe
- 성공/실패 피드백
- 목록 페이지 하단 + 상세 페이지 하단에 배치
```

### 6단계: 카카오 공유 (Day 8)

```
- Kakao JavaScript SDK 연동
- KakaoShare 컴포넌트
- 공유 시 공고 카드 형태의 미리보기
  - 제목: 사업명
  - 설명: AI 요약 2줄
  - 이미지: 동네정보 기본 OG 이미지
  - 링크: 해당 공고 상세 페이지
```

---

## MVP 이후 (Phase 2)

```
- 의견 작성/목록
- 찬반 투표 (찬성/반대 버튼)
- 이슈 추적 ([🔔 이 이슈 추적하기] 버튼)
- 다크 모드
- PWA (홈 화면 추가)
- 검색 기능
```
