# CLAUDE.md — 동네정보 프로젝트 가이드

## 프로젝트 개요

우리 동네 행정 공고를 자동 수집하고, AI가 쉽게 설명해서 주민에게 알려주고, 주민 의견을 모아주는 시빅테크 플랫폼.
상세 기획은 `docs/` 참고.

## 기술 스택

- 프론트: Next.js 14+ (App Router) + Tailwind CSS (`web/`)
- 백엔드: Go 1.23+
- 프론트: Next.js (추후)
- DB: PostgreSQL (Supabase)
- AI: Claude API
- 알림: 카카오 알림톡
- 배포: Railway (Go) + Vercel (Next.js)

## 프로젝트 구조 (레이어 기반)

```
dongne-info/
├── cmd/
│   ├── server/         # API 서버 진입점
│   │   └── main.go
│   └── crawl/          # 크롤러 진입점
│       └── main.go
├── config/             # 설정 관리
│   └── config.go
├── crawler/            # 데이터 수집 (서울시 API, 구청 크롤링)
│   ├── rebuild.go      # 재개발/재건축
│   ├── cityplan.go     # 도시계획 결정고시
│   └── district.go     # 자치구 공고
├── model/              # 데이터 구조체 (DB 모델)
│   ├── announcement.go
│   ├── subscriber.go
│   ├── reaction.go
│   └── errors.go       # 커스텀 에러 타입
├── repository/         # DB 접근 (sqlx)
│   ├── announcement.go
│   ├── subscriber.go
│   └── reaction.go
├── service/            # 비즈니스 로직
│   ├── announcement.go
│   ├── subscriber.go
│   ├── notification.go
│   └── summary.go      # Claude API 요약
├── api/                # HTTP 핸들러 (Gin)
│   ├── router.go
│   ├── announcement_handler.go
│   ├── subscriber_handler.go
│   ├── reaction_handler.go
│   └── middleware/
│       └── cors.go
├── docs/               # 기획 문서
│   ├── README.md
│   ├── BUSINESS_MODEL.md
│   ├── ARCHITECTURE.md
│   └── DEVELOPMENT.md
├── migrations/         # DB 마이그레이션
│   └── 001_init.sql
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── CLAUDE.md
```

## 코딩 컨벤션

### 네이밍

```
파일명:    snake_case.go (announcement_handler.go)
패키지명:  소문자 단일 단어 (crawler, api, model)
구조체:    PascalCase (Announcement, Subscriber)
함수:      PascalCase(공개), camelCase(비공개)
변수:      camelCase
상수:      PascalCase or ALL_CAPS
DB 컬럼:   snake_case (source_id, created_at)
API URL:   kebab-case (/api/announcements/:id/reactions)
```

### 에러 처리: 커스텀 에러 타입

```go
// model/errors.go

type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found (id=%s)", e.Resource, e.ID)
}

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s - %s", e.Field, e.Message)
}

type ExternalAPIError struct {
    Service string
    Err     error
}

func (e *ExternalAPIError) Error() string {
    return fmt.Sprintf("external API error (%s): %v", e.Service, e.Err)
}

func (e *ExternalAPIError) Unwrap() error {
    return e.Err
}
```

핸들러에서 에러 타입에 따라 HTTP 상태 코드를 분기:
```go
var notFound *model.NotFoundError
if errors.As(err, &notFound) {
    c.JSON(http.StatusNotFound, ...)
    return
}
```

### DB 접근: sqlx

```go
// repository/announcement.go

type AnnouncementRepository struct {
    db *sqlx.DB
}

func NewAnnouncementRepository(db *sqlx.DB) *AnnouncementRepository {
    return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) FindByDistrict(ctx context.Context, district string) ([]model.Announcement, error) {
    var announcements []model.Announcement
    query := `SELECT * FROM announcements WHERE district = $1 ORDER BY created_at DESC`
    if err := r.db.SelectContext(ctx, &announcements, query, district); err != nil {
        return nil, fmt.Errorf("공고 조회 실패 (district=%s): %w", district, err)
    }
    return announcements, nil
}
```

- SQL은 항상 명시적으로 작성 (ORM 금지)
- context.Context는 항상 첫 번째 파라미터로
- 구조체 태그로 DB 컬럼 매핑: `db:"column_name"`

### HTTP: Gin

```go
// api/router.go

func SetupRouter(h *Handlers) *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        api.GET("/announcements", h.Announcement.List)
        api.GET("/announcements/:id", h.Announcement.Get)
        api.POST("/subscribe", h.Subscriber.Create)
        api.POST("/announcements/:id/react", h.Reaction.Create)
        api.GET("/announcements/:id/reactions", h.Reaction.GetByAnnouncement)
    }

    return r
}
```

### 설정: 구조체 + 환경변수

```go
// config/config.go

type Config struct {
    Port           string `env:"PORT" envDefault:"8080"`
    DatabaseURL    string `env:"DATABASE_URL,required"`
    SeoulAPIKey    string `env:"SEOUL_API_KEY,required"`
    AnthropicKey   string `env:"ANTHROPIC_API_KEY,required"`
    KakaoAPIKey    string `env:"KAKAO_API_KEY"`
    Environment    string `env:"ENVIRONMENT" envDefault:"development"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("설정 로드 실패: %w", err)
    }
    return cfg, nil
}
```

### 로깅: slog

```go
slog.Info("크롤링 시작", "district", district, "source", "upisRebuild")
slog.Error("크롤링 실패", "district", district, "error", err)
```

- 항상 구조화된 키-값 쌍으로 로깅
- fmt.Println이나 log.Println 사용 금지

### 의존성 주입 패턴

```go
// 레이어 간 의존성은 생성자 주입

repo := repository.NewAnnouncementRepository(db)
svc := service.NewAnnouncementService(repo)
handler := api.NewAnnouncementHandler(svc)
```

- 인터페이스는 사용하는 쪽에서 정의 (Go 관례)
- 전역 변수/싱글턴 금지

## 테스트 규칙

### 테스트는 꼼꼼하게

- 모든 service 레이어 함수에 테스트 작성
- repository는 통합 테스트 (테스트 DB 사용)
- handler는 httptest로 E2E 테스트
- 크롤러는 모킹된 HTTP 응답으로 테스트

### 테이블 드리븐 + 서브테스트

```go
func TestAnnouncementService_FindByDistrict(t *testing.T) {
    tests := []struct {
        name     string
        district string
        wantErr  bool
        wantLen  int
    }{
        {
            name:     "강남구 공고 조회",
            district: "강남구",
            wantErr:  false,
            wantLen:  5,
        },
        {
            name:     "빈 구 이름",
            district: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### 테스트 파일 위치

```
service/
├── announcement.go
├── announcement_test.go    # 같은 패키지 내
```

## 주석 규칙

### 모든 공개 구조체/함수에 상세 한글 주석 필수

```go
// Announcement 는 서울시 행정 공고 데이터를 나타내는 구조체.
//
// 서울시 열린데이터광장 API에서 수집한 재개발/재건축/도시계획 공고를 저장한다.
// crawler 패키지에서 생성하고, service/api 레이어에서 조회/응답에 사용한다.
//
// 필드:
//   - District: "강남구", "마포구" 등 서울시 자치구명
//   - Type: "재개발", "재건축", "도시계획" 중 하나
//   - Action: "신설", "변경", "폐지", "기타" 중 하나
//   - Summary: Claude API가 생성한 쉬운 설명 (nil이면 아직 요약 안 됨)
//   - Stage: 재건축 진행 단계 (예: "정비구역 지정 - 7단계 중 2단계")
type Announcement struct { ... }
```

### 주석에 포함해야 할 내용
1. **정의**: 이 구조체/함수가 뭔지
2. **사용법**: 어떤 레이어/패키지에서 어떻게 쓰는지
3. **사용 시점**: 언제 호출하는지, 어떤 조건에서 쓰는지
4. **주의사항**: 있으면 작성 (nil 체크, 동시성 등)

### 비공개 함수도 로직이 복잡하면 주석 추가

```go
// classifyAction 은 서울시 API의 RPT_TYPE 필드를 분석해서
// "신설", "변경", "폐지", "기타" 중 하나로 분류한다.
//
// RPT_TYPE에 "신설", "지정", "설립" 등의 키워드가 포함되면 "신설"로 판단.
// 키워드가 없으면 "기타"로 분류한다.
func classifyAction(rptType string) string { ... }
```

## 코드 작성 원칙

1. **한국어 주석, 영어 코드**: 변수/함수명은 영어, 주석과 에러 메시지는 한국어 OK
2. **주석은 상세하게**: 구조체/함수 위에 정의 + 사용법 + 시점 + 주의사항
3. **코드 수정 시 주석도 반드시 동기화**: 필드 추가/삭제/변경하면 주석도 같이 업데이트
4. **Simple first**: 추상화보다 명시적인 코드. 3번 반복되면 그때 추상화
5. **레이어 경계 지키기**: handler → service → repository 순서. 역방향 의존 금지
6. **context.Context 전파**: 모든 DB/외부 API 호출에 context 전달
7. **에러는 래핑해서 올리기**: 어디서 터졌는지 추적 가능하게
8. **매직 넘버 금지**: 상수로 정의
9. **환경별 분기 최소화**: config로 처리, 코드에 if production 같은 분기 넣지 않기

---

## 프론트엔드 코딩 컨벤션 (web/)

### 네이밍

```
컴포넌트 파일:  PascalCase.tsx (AnnouncementCard.tsx)
유틸/라이브러리: camelCase.ts (api.ts, utils.ts)
타입 파일:      camelCase.ts (types/index.ts)
페이지:         page.tsx (Next.js App Router 규칙)
레이아웃:       layout.tsx (Next.js App Router 규칙)
CSS 클래스:     Tailwind 유틸리티 클래스 (커스텀 CSS 최소화)
환경변수:       NEXT_PUBLIC_ 접두사는 클라이언트 노출용만, 그 외는 서버 전용
```

### 주석: "타입이 곧 문서다"

React/TypeScript에서는 코드 자체가 문서화되어야 한다.
Go처럼 모든 곳에 상세 JSDoc을 다는 것이 아니라, 수준에 맞게 분리한다.

**단순 컴포넌트 (TypeBadge, AreaChange 등):**
JSDoc 없이 Props 인터페이스의 속성 옆에 짧은 주석으로 충분.

```tsx
interface TypeBadgeProps {
  type: string;     // "재개발", "재건축", "도시계획"
  variant: "type" | "action"; // type=유형 뱃지, action=조치 뱃지
}

// 공고 유형/조치를 색상 뱃지로 표시
export default function TypeBadge({ type, variant }: TypeBadgeProps) { ... }
```

**복잡한 컴포넌트 (StageBar, InterestButton 등):**
비즈니스 로직이나 조건부 렌더링이 있으면 상세 JSDoc.

```tsx
/**
 * StageBar — 재건축 7단계 진행 상황을 프로그레스 바로 표시.
 *
 * 조건부 렌더링:
 *   - "2/7단계 - 정비구역지정" → 프로그레스 바 (2/7 채움) + 텍스트
 *   - "확인 필요" → 텍스트만
 *   - null/undefined → 렌더링 안 함 (return null)
 *
 * @param stage - "2/7단계 - 정비구역지정" 형태 또는 null
 */
```

**페이지 (page.tsx):**
해당 페이지의 목적과 API 데이터 소스를 상단에 명시.

```tsx
/**
 * 강남구 공고 목록 페이지 (/gangnam)
 *
 * 데이터: GET /api/announcements?district=강남구
 * 서버 컴포넌트 (SSR, SEO용)
 */
```

### 서버/클라이언트 컴포넌트 구분

```
서버 컴포넌트 (기본값):
  - 데이터 페칭, SEO가 필요한 것
  - 공고 목록, 공고 상세, 레이아웃
  - "use client" 없으면 서버 컴포넌트

클라이언트 컴포넌트 ("use client"):
  - 유저 인터랙션이 있는 것만
  - 관심 버튼, 구독 폼, 필터 탭, 카카오 공유
  - useState, useEffect, onClick 등이 필요한 것만
```

```
규칙:
- 기본은 서버 컴포넌트. "use client"는 진짜 필요한 곳에만.
- 서버 컴포넌트 안에 클라이언트 컴포넌트를 자식으로 넣는 구조.
- 절대로 페이지(page.tsx) 전체를 "use client"로 만들지 않는다. SEO 죽음.
```

### API 호출 패턴

API 프록시 없이 Go API에 직접 통신. MVP에서는 속도가 우선.

```tsx
// 서버 컴포넌트에서 (page.tsx):
// Go API를 직접 호출 (서버 사이드라 CORS 없음)
const res = await fetch(`${process.env.GO_API_URL}/api/announcements?district=강남구`);
const data = await res.json();

// 클라이언트 컴포넌트에서:
// NEXT_PUBLIC_ 환경변수로 Go API 직접 호출 (Go에 CORS 미들웨어 추가)
const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/announcements/${id}/react`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ type: "interest", session_id: sessionId }),
});
```

환경변수:
```
GO_API_URL=http://localhost:8080            # 서버 컴포넌트용 (비노출)
NEXT_PUBLIC_API_URL=http://localhost:8080    # 클라이언트 컴포넌트용 (브라우저 노출)
```

### 상태 피드백 (로딩/에러) 패턴

유저가 행동했을 때 반드시 피드백을 줘야 한다.

**로딩:**
```tsx
// 버튼 내 스피너 + disabled
<button disabled={isLoading} onClick={handleClick}>
  {isLoading ? <Spinner /> : "📌 관심 있어요"}
</button>
```

**에러:**
```tsx
// 컴포넌트 단위 인라인 에러 메시지. 전역 ErrorBoundary보다 가벼움.
{error && <p className="text-red-500 text-sm mt-1">{error}</p>}
```

**낙관적 업데이트 (관심 버튼 등):**
```tsx
// 서버 응답 전에 UI를 먼저 반영. 실패하면 롤백.
const handleInterest = async () => {
  setCount(prev => prev + 1);  // 즉시 UI 반영
  try {
    await fetch(...);
  } catch {
    setCount(prev => prev - 1);  // 실패 시 롤백
    setError("다시 시도해주세요.");
  }
};
```

**서버 컴포넌트 에러:**
```tsx
import { notFound } from "next/navigation";

const res = await fetch(`${API_URL}/api/announcements/${id}`);
if (!res.ok) {
  notFound(); // Next.js 404 페이지 표시
}
```

### 폼 처리: Simplicity First

```
규칙: useState + 기본 HTML5 Validation을 기본으로 한다.
확장: 입력 필드 5개 이상이거나 복잡한 검증이 필요한 경우에만 React Hook Form 도입 검토.
```

```tsx
// ✅ MVP 폼 패턴
const [email, setEmail] = useState("");
const [isSubmitting, setIsSubmitting] = useState(false);

<form onSubmit={handleSubmit}>
  <input
    type="email"
    required                    // HTML5 기본 검증
    value={email}
    onChange={(e) => setEmail(e.target.value)}
  />
  <button type="submit" disabled={isSubmitting}>
    {isSubmitting ? "구독 중..." : "구독하기"}
  </button>
</form>
```

### 접근성 (a11y): 공익 서비스의 본질

타겟 층(30~50대)을 고려할 때 접근성은 '선택'이 아닌 '필수'.

```
규칙:
1. 이미지: 모든 <img>, <Image>에 의미 있는 alt 필수
2. 버튼: 아이콘만 있는 버튼(닫기 X, 공유 등)은 반드시 aria-label 부여
3. 색상: 텍스트와 배경의 대비를 높여 노안/시력 약화 유저 배려 (WCAG AA 기준)
4. 키보드: 탭 순서가 자연스럽게 흐르도록. 커스텀 버튼도 키보드로 동작해야 함
5. 터치 영역: 모바일 터치 타겟 최소 44px × 44px
```

```tsx
// ✅ 좋은 예
<button aria-label="카카오톡으로 공유" onClick={handleShare}>
  <ShareIcon />
</button>

<Image src="/logo.png" alt="동네정보 로고" width={120} height={40} />

// ❌ 나쁜 예
<button onClick={handleShare}><ShareIcon /></button>  // aria-label 없음
<img src="/logo.png" />                                // alt 없음
```

### Tailwind 사용 규칙

```
- 인라인 Tailwind 클래스 사용 (별도 CSS 파일 만들지 않음)
- 반복되는 스타일 조합은 컴포넌트로 추출 (유틸 클래스 만들지 않음)
- 모바일 퍼스트: 기본 스타일이 모바일, sm:/md:/lg:로 확장
- 색상은 Tailwind 기본 팔레트 사용, 커스텀 색상은 tailwind.config.ts에 정의
```

```tsx
// ✅ 좋은 예: 모바일 퍼스트
<div className="p-4 md:p-6 lg:p-8">

// ❌ 나쁜 예: 데스크탑 퍼스트
<div className="p-8 sm:p-4">
```

### 파일 당 하나의 컴포넌트

```
✅ components/announcement/AnnouncementCard.tsx → export default function AnnouncementCard
✅ components/announcement/TypeBadge.tsx → export default function TypeBadge

❌ components/announcement/index.tsx에 여러 컴포넌트 몰아넣기
```
