package model

import "time"

// Reaction 은 공고에 대한 주민 반응(관심 표시, 찬성, 반대)을 나타내는 구조체.
//
// 주민이 공고 카드에서 [관심 있어요], [찬성], [반대] 버튼을 누르면 생성된다.
// 비로그인 유저도 session_id 기반으로 반응할 수 있다 (중복 방지).
// 같은 유저가 같은 공고에 같은 타입의 반응을 중복으로 하면 DB UNIQUE 제약에 의해 무시된다.
//
// 필드:
//   - AnnouncementID: 반응 대상 공고 ID
//   - Type: 반응 유형 ("interest", "agree", "disagree" 중 하나)
//   - SessionID: 비로그인 유저 식별용 세션 ID (쿠키 기반, nil이면 로그인 유저)
//   - SubscriberID: 로그인한 구독자 ID (nil이면 비로그인 유저)
type Reaction struct {
	ID             string    `db:"id" json:"id"`
	AnnouncementID string    `db:"announcement_id" json:"announcement_id"`
	Type           string    `db:"type" json:"type"`
	SessionID      *string   `db:"session_id" json:"session_id,omitempty"`
	SubscriberID   *string   `db:"subscriber_id" json:"subscriber_id,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// ReactionSummary 는 특정 공고에 대한 반응 집계 결과.
//
// GET /api/announcements/:id/reactions 응답에 사용된다.
// repository에서 SQL COUNT 집계로 생성한다.
//
// 사용 예 (응답):
//
//	{"interest": 847, "agree": 231, "disagree": 56}
type ReactionSummary struct {
	Interest int `json:"interest" db:"interest"`
	Agree    int `json:"agree" db:"agree"`
	Disagree int `json:"disagree" db:"disagree"`
}

// CreateReactionRequest 는 반응 등록 API 요청 바디 구조체.
//
// POST /api/announcements/:id/react 핸들러에서 사용한다.
// 비로그인 유저는 SessionID를 쿠키에서 생성하여 전달한다.
//
// 필드:
//   - Type: "interest", "agree", "disagree" 중 하나 (필수)
//   - SessionID: 비로그인 유저 식별용 (선택)
type CreateReactionRequest struct {
	Type      string  `json:"type" binding:"required,oneof=interest agree disagree"`
	SessionID *string `json:"session_id"`
}
