package model

import "time"

// Comment 는 공고에 대한 주민 의견(한줄 댓글)을 나타내는 구조체.
//
// 주민이 공고 상세 페이지에서 의견을 작성하면 생성된다.
// Reaction(관심/찬반)과 달리 로그인(구독자)만 작성 가능하다.
// Phase 2에서 의견 수가 N개 이상이면 "주민 의견 리포트"를 자동 생성하는 데 사용된다.
//
// 필드:
//   - AnnouncementID: 의견 대상 공고 ID
//   - SubscriberID: 작성자 구독자 ID (로그인 필수)
//   - Body: 의견 본문 (최대 500자)
type Comment struct {
	ID             string    `db:"id" json:"id"`
	AnnouncementID string    `db:"announcement_id" json:"announcement_id"`
	SubscriberID   string    `db:"subscriber_id" json:"subscriber_id"`
	Body           string    `db:"body" json:"body"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// CreateCommentRequest 는 의견 작성 API 요청 바디 구조체.
//
// POST /api/announcements/:id/comments 핸들러에서 사용한다.
// 로그인한 구독자만 작성 가능하므로 SubscriberID가 필수이다.
//
// 필드:
//   - Body: 의견 본문 (필수, 1~500자)
//   - SubscriberID: 작성자 구독자 ID (필수)
type CreateCommentRequest struct {
	Body         string `json:"body" binding:"required,min=1,max=500"`
	SubscriberID string `json:"subscriber_id" binding:"required"`
}
