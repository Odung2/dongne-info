# Architecture

## 시스템 구성도

```
[공공 API / 크롤링]
  ├── ★ 서울 열린데이터광장 - upisRebuild (재개발/정비사업) ← 1단계, 완료
  ├── ★ 서울 열린데이터광장 - upisCityplan (도시계획 결정고시) ← 1단계
  ├── ★ 서울시 자치구 공고 RSS/크롤링 (구청 홈페이지 고시·공고) ← 1단계
  ├── 정비사업 정보몽땅 (cleanup.seoul.go.kr) ← 2단계
  ├── 국토부 실거래가 API (시세 참고 데이터) ← 2단계
  └── 국회지방의회의정포털 (구의회 회의록) ← 3단계(선거 특집)

       ↓ Go Crawler (매일 자동 실행)

[PostgreSQL - Supabase]
  ├── announcements (공고 테이블)
  ├── districts (구 정보)
  ├── subscribers (구독자)
  ├── reactions (관심 표시 + 찬반)
  ├── comments (주민 의견)
  └── notification_logs (알림 발송 이력)

       ↓

[Go API Server]
  ├── REST API
  ├── Claude API 연동 (AI 쉬운 설명)
  └── 카카오 알림톡 발송

       ↓

[Next.js 웹 프론트]
  ├── 구별 공고 목록
  ├── AI 해석 상세페이지
  ├── 주민 참여 (관심 표시, 의견 남기기)
  ├── 이메일/카카오 알림 구독
  └── 선거 특집 페이지 (6월)

[배포]
  ├── Go 서버: Railway or Render
  ├── Next.js: Vercel
  └── DB: Supabase (PostgreSQL)
```

---

## 기술 스택

| 영역 | 기술 | 이유 |
|---|---|---|
| 백엔드 | Go | 빠름, 크롤러에 적합 |
| 프론트 | Next.js | SEO 중요 (구글 검색 유입) |
| DB | PostgreSQL (Supabase) | 클라우드, 무료 플랜 |
| AI | Claude API | 공고 해석, 요약 |
| 알림 | 카카오 알림톡 | 오픈율 높음 |
| 배포 | Vercel + Railway | 무료 시작 가능 |

---

## DB 스키마

### announcements (공고)
```sql
CREATE TABLE announcements (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  district    VARCHAR(20) NOT NULL,       -- 강남구, 마포구
  type        VARCHAR(20) NOT NULL,       -- 재개발, 재건축, 도시계획
  action      VARCHAR(10) NOT NULL,       -- 신설, 변경, 폐지
  title       TEXT NOT NULL,              -- 원본 제목
  location    TEXT,                       -- 위치
  summary     TEXT,                       -- Claude AI 쉬운 설명 (행정용어 → 일반어)
  stage       VARCHAR(50),               -- 진행 단계 (예: "정비구역 지정 - 7단계 중 2단계")
  related     TEXT,                       -- 관련 과거 사례 또는 유사 지역 정보
  source_url  TEXT,                       -- 원본 링크
  source_id   VARCHAR(100) UNIQUE,        -- 중복 방지용 원본 ID
  created_at  TIMESTAMP DEFAULT NOW(),
  notified_at TIMESTAMP                   -- 알림 발송 시각
);
```

### subscribers (구독자)
```sql
CREATE TABLE subscribers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  contact     VARCHAR(100) NOT NULL,      -- 이메일 or 카카오
  type        VARCHAR(10) NOT NULL,       -- email, kakao
  districts   TEXT[],                     -- 구독 구 목록
  created_at  TIMESTAMP DEFAULT NOW(),
  active      BOOLEAN DEFAULT TRUE
);
```

### reactions (관심 표시 / 찬반)
```sql
CREATE TABLE reactions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  type            VARCHAR(10) NOT NULL,   -- interest, agree, disagree
  session_id      VARCHAR(100),           -- 비로그인 유저 식별 (쿠키 기반)
  subscriber_id   UUID REFERENCES subscribers(id), -- 로그인 유저 (선택)
  created_at      TIMESTAMP DEFAULT NOW(),
  UNIQUE(announcement_id, session_id, type)
);
```

### comments (주민 의견)
```sql
CREATE TABLE comments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  subscriber_id   UUID REFERENCES subscribers(id),
  body            TEXT NOT NULL,
  created_at      TIMESTAMP DEFAULT NOW()
);
```

### notification_logs (알림 이력)
```sql
CREATE TABLE notification_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  subscriber_id   UUID REFERENCES subscribers(id),
  sent_at         TIMESTAMP DEFAULT NOW(),
  status          VARCHAR(10)             -- success, fail
);
```

---

## API 엔드포인트

```
GET  /api/announcements?district=강남구&type=재개발
GET  /api/announcements/:id
POST /api/subscribe
GET  /api/districts

POST /api/announcements/:id/react        -- 관심 표시 / 찬반
GET  /api/announcements/:id/reactions     -- 반응 집계
POST /api/announcements/:id/comments      -- 의견 남기기
GET  /api/announcements/:id/comments      -- 의견 목록
```

---

## 크롤러 플로우

```
1. 매일 오전 9시 cron 실행
2. 서울시 API 호출 (강남구, 마포구) — 3개 데이터 소스
3. source_id로 중복 체크
4. 새 공고면 DB 저장
5. Claude API로 쉬운 설명 + 진행 단계 + 관련 사례 생성
6. 구독자에게 카카오 알림톡 발송
7. notification_logs에 기록
```

---

## 알림 설계 원칙

오픈율이 핵심 지표이므로, 알림 자체의 설계가 BM에 직결된다.

```
❌ 구 단위 알림 (오픈율 낮음)
"강남구 새 공고 2건이 등록되었습니다"
→ "뭔 소리야" → 안 열음

✅ 이슈 단위 알림 (오픈율 높음)
"은마아파트 재건축 관련 새 공고가 나왔어요"
→ "우리 아파트!" → 반드시 열음
```

### 알림 우선순위

```
1순위: 추적 중인 이슈에 새 공고 → 무조건 발송 (가장 높은 오픈율)
2순위: 구독한 구의 관심 높은 이슈 → 주 2~3회 발송
3순위: 주간 요약 → 주 1회 (놓친 공고 정리)
```

### 알림 포맷

```
[카카오 알림톡]
📌 추적 중인 이슈에 새 소식이 있어요

은마아파트 재건축
━━━━━━━━━━━━━━
새 공고: "정비계획 변경 결정 고시"
쉬운 설명: 재건축 설계 일부가 바뀌었어요
진행 단계: ███░░░░ 3/7단계 (변동 없음)
관심: 847명 (+52 이번 주)

→ 자세히 보기
```

핵심: **"내 아파트/내가 추적하는 이슈"로 개인화해야 오픈율이 나온다.**
"강남구 소식" 같은 범용 알림은 뉴스레터와 다를 게 없고, 오픈율도 20%대로 떨어진다.

---

## 참여 → 액션 루프

핵심: 참여하면 실제로 뭔가가 움직이고, 그 결과를 추적할 수 있어야 한다.

```
공고 수집 → AI 쉬운 설명 → 참여 → 액션 → 결과 → 팔로우업 → 재참여
                              ↑                              │
                              └──────────────────────────────┘
```

### Phase 1 (MVP): 관심 표시 + 이슈 추적

```
[참여]
  공고 카드 → [📌 관심 있어요] 버튼
  → 숫자 표시: "이 공고에 324명이 관심 있어요"
  → 비로그인도 가능 (session 기반)

[이슈 추적]
  [🔔 이 이슈 추적하기] 버튼
  → 해당 이슈에 새 공고/진전이 있으면 알림
  → 재건축은 수년 걸림 → 장기 리텐션 장치

  추적 중인 이슈:
  ────────────────
  은마아파트 재건축  ███░░░░ 3/7단계  📌 추적 중 (412명)
  성산시영 정비계획  ██░░░░░ 2/7단계  📌 추적 중 (231명)
```

### Phase 2: 의견 수집 + 구의회 전달

```
[의견 수집]
  찬반 + 한줄 의견
  → "주민 의견 리포트" 자동 생성

[액션: 구의회 전달]
  관심 N명 넘으면 → 구의회에 주민 의견 공식 전달
  → "전달 완료!" 결과 공유
  → "847명이 관심, 56건의 의견이 접수되었습니다"

[팔로우업]
  전달 후 구의회 논의 상황 추적 → 알림
  → "구의회 안건 상정됨" / "논의 결과: ~" / "최종 결과: ~"
  → 이 알림이 재방문과 재참여를 만듦
```

### Phase 3: 구청 질문 대행

```
[액션: 구청 대표 질문]
  "이 공고에 대해 구청에 질문하기"
  → 같은 질문 N명 이상이면 대표 질문으로 구청에 전달
  → 혼자 민원 넣기 부담 → "23명이 같은 질문" → 심리적 안전

[결과]
  구청 답변 도착 → AI가 쉽게 재해석 → 질문자 전원에게 알림
  → 구청 답변 자체가 새 콘텐츠가 됨
```

### 효능감 루프가 만드는 것

```
참여 → "진짜 전달됐다" (효능감) → 다음에도 참여
추적 → "그 건 어떻게 됐지?" (호기심) → 재방문
공유 → "500명 넘어야 전달돼" (동기) → 이웃에게 바이럴
답변 → 구청/구의회 응답이 콘텐츠 (자동 공급) → 공고 없는 날에도 콘텐츠
```

---

## 참여 → 액션 루프 객관적 평가

### 강점

**1. 효능감 루프는 검증된 메커니즘**
Change.org(서명→전달→결과), 국민청원(20만→정부답변) 등
"내가 한 행동이 결과를 만들었다" → 재참여. 행동심리학적으로 확실한 동력.

**2. 팔로우업 = 자연스러운 리텐션**
재건축 평균 10~15년. 한 번 추적 누르면 수년간 알림 받을 이유.
뉴스레터는 매번 새 콘텐츠를 만들어야 하지만,
동네정보는 기존 이슈의 진전 자체가 콘텐츠.

**3. 바이럴 내장**
"N명 넘으면 구의회 전달" → 자연스러운 공유 동기.

### 약점

**1. 구의회 전달 — 실제로 효과 있냐?**
구의회가 "아 그래요" 하고 무시하면 효능감 0.
법적 접수 의무 없음. 전달은 쉬운데 "바뀌었다"는 다른 문제.
→ 대안: 직접 전달보다 "주민 관심 리포트" 공개가 더 강력할 수 있음.
   언론이 가져가면 자연스럽게 압력이 됨.

**2. 구청 질문 대행 — 행정적 허들**
민원은 개인 실명으로 넣어야 함. "대행"이 행정 절차상 가능한지 불확실.
→ 대안: 대행보다 "민원 넣기 가이드 + 질문 템플릿 AI 자동 생성".
   유저가 직접 제출, "저도 넣었어요" 체크로 숫자 표시.

**3. 초기 숫자 문제**
유저 50명일 때 "관심 3명, 500명까지 497명 남음" → 한심해 보임.
→ 대안: 초기엔 임계점 없이 "주민 의견 리포트 매주 자동 생성".

**4. 팔로우업을 누가 하냐?**
"구의회 안건 상정됐다" 정보를 수동으로 추적? 혼자 운영인데 매주 가능?
→ 대안: 같은 location/type의 새 공고 나오면 자동 연결. 크롤러가 이미 하는 일.
   구의회 회의록 연동은 나중에.

### 기능별 현실성 평가

| 기능 | 가치 | 현실성 | 시점 |
|---|---|---|---|
| 관심 표시 + 숫자 | 중 | 높음 | MVP |
| 이슈 추적 + 알림 | 높음 | 높음 | MVP |
| 주민 관심 리포트 공개 | 높음 | 높음 | Phase 2 |
| 구의회 전달 | 높음 | 낮음 | Phase 3 |
| 구청 질문 대행 | 중 | 낮음 | Phase 3 |

---

## "이슈 추적"은 진짜 리텐션을 만드냐? — 벤치마크 비교 분석

### 핵심 질문

이슈 추적이 "매일 들어올 이유"가 되냐?
솔직히 말하면: **안 된다.**

```
은마아파트 재건축 추적 중
→ 새 공고가 나오는 빈도: 수개월에 1번
→ 알림: "3개월 만에 새 공고 나왔어요"
→ 들어와서 읽음 → 다시 수개월 대기
```

이건 "매일 들어오는 서비스"가 아니라 "가끔 알림 오는 서비스"야.

### 벤치마크와 비교

```
당근마켓:   중고거래 → 매일 새 매물 → 매일 들어옴        (DAU 높음)
Nextdoor:  이웃 소식 → 올라올 수도 있고 없을 수도 → 가끔 (DAU 낮음 → 실패)
동네정보:   공고 추적 → 수개월에 1번 알림 → 극히 가끔     (DAU ???)
```

동네정보의 이슈 추적은 **Nextdoor보다도 방문 빈도가 낮을 수 있어.**

### 그러면 이슈 추적은 쓸모없냐?

아니. 역할이 다른 거야.

```
이슈 추적의 진짜 가치:
- DAU를 만드는 기능이 아님
- "이탈 방지" 기능임
- 한번 온 유저가 서비스를 삭제/잊어버리지 않게 하는 장치
- 수개월 후에도 "아 그 서비스" 하고 돌아오게 만드는 장치
```

비유하면:
```
당근 = 매일 가는 편의점 (DAU)
동네정보 = 가끔 가지만 절대 해지 안 하는 보험 앱 (리텐션)
```

보험 앱은 매일 안 열어도 "내 보험 상태 추적" 때문에 삭제 안 해.
동네정보도 매일 안 열어도 "내 아파트 재건축 추적" 때문에 구독 해지 안 해.

### 진짜 문제: 그러면 DAU는 뭘로 만드냐?

이슈 추적만으로는 DAU가 안 나오고, DAU가 없으면 광고가 안 팔려.
벤치마크에서 배운 교훈:

```
DAU를 만드는 건:
- 당근: 매일 바뀌는 콘텐츠 (새 매물)
- The Skimm: 매일 오는 뉴스레터 (습관)
- 커뮤니티: 다른 사람의 새 글 (UGC)
```

동네정보에서 "매일 바뀌는 것"은 뭐가 있냐:
```
공고: 주 2~3건 → 매일은 아님
추적 알림: 수개월에 1번 → 매일 아님
관심 숫자: 조금씩 바뀜 → 약함
의견/댓글: 유저가 써야 바뀜 → 초기엔 없음
```

**솔직한 결론: 동네정보는 구조적으로 DAU 서비스가 아니다.**

### 그러면 어떻게 하냐?

DAU를 억지로 만들려 하지 말고, **이 서비스의 자연스러운 리듬에 맞추는 게 맞아.**

```
동네정보의 자연스러운 리듬:
- 주 2~3회 알림 (공고 기반)
- 월 1~2회 이슈 추적 알림
- 선거 시즌 집중 (일시 폭발)

이건 "매일 여는 앱"이 아니라 "알림 올 때 여는 앱"
= 뉴스레터에 가까운 리듬
= 그리고 그게 나쁜 게 아님
```

The Skimm도 매일 앱을 여는 게 아니라 **매일 메일이 오니까 여는 거야.**
동네정보도 **알림이 올 때 열리는 서비스**로 설계하면,
DAU는 낮아도 **알림 오픈율이 높으면 광고가 팔린다.**

```
The Skimm: 구독 700만, 오픈율 35% → 매회 245만 도달 → 광고 잘 팔림
동네정보:  구독 5,000, 오픈율 50% (내 동네 건이니까) → 매회 2,500 도달
          → 로컬 업체 광고로서는 충분한 도달
```

### 최종 정리

```
이슈 추적의 역할:
- DAU를 만드는 기능 ❌
- 구독 유지 + 이탈 방지 ✅
- 장기 리텐션 (수년) ✅
- 효능감의 기반 ✅

DAU 대신 노려야 할 것:
- 높은 알림 오픈율 (타겟 정밀도로 가능)
- 알림 올 때마다 참여 (관심/의견/추적)
- 참여 → 액션 루프로 "다음 알림도 열어볼 이유" 생성

이 서비스는 "매일 여는 앱"이 아니라 "알림 올 때 반드시 여는 앱"이다.
그리고 그게 이 서비스에 맞는 리듬이다.
```
