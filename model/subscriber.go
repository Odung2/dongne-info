package model

import (
	"time"

	"github.com/lib/pq"
)

// Subscriber 는 공고 알림을 구독한 사용자를 나타내는 구조체.
//
// 주민이 이메일 또는 카카오를 통해 특정 구의 공고 알림을 구독하면 생성된다.
// 새 공고가 수집되면 notification 서비스가 이 테이블에서 해당 구 구독자를 조회하여
// 알림을 발송한다.
//
// 필드:
//   - Contact: 연락처 (이메일 주소 또는 카카오 식별자)
//   - Type: 연락처 유형 ("email" 또는 "kakao")
//   - Districts: 구독 중인 자치구 목록 (예: ["강남구", "마포구"])
//   - Active: 구독 활성 상태 (false면 알림 발송하지 않음)
type Subscriber struct {
	ID        string         `db:"id" json:"id"`
	Contact   string         `db:"contact" json:"contact"`
	Type      string         `db:"type" json:"type"`
	Districts pq.StringArray `db:"districts" json:"districts"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	Active    bool           `db:"active" json:"active"`
}

// CreateSubscriberRequest 는 구독 등록 API 요청 바디 구조체.
//
// POST /api/subscribe 핸들러에서 JSON 바디를 파싱할 때 사용한다.
// Gin의 binding 태그로 필수 필드 및 값 범위를 검증한다.
//
// 사용 예:
//
//	{
//	  "contact": "user@example.com",
//	  "type": "email",
//	  "districts": ["강남구"]
//	}
//
// 필드:
//   - Contact: 연락처 (필수)
//   - Type: "email" 또는 "kakao" (필수, 다른 값이면 400 에러)
//   - Districts: 구독할 자치구 목록 (필수, 최소 1개)
type CreateSubscriberRequest struct {
	Contact   string   `json:"contact" binding:"required"`
	Type      string   `json:"type" binding:"required,oneof=email kakao"`
	Districts []string `json:"districts" binding:"required,min=1"`
}
