package repository

import (
	"context"
	"fmt"

	"dongne-info/model"

	"github.com/jmoiron/sqlx"
)

// ReactionRepository 는 reactions 테이블에 대한 CRUD를 담당한다.
//
// 주민의 관심 표시/찬반 반응을 저장하고 집계한다.
// service.ReactionService에서 사용한다.
type ReactionRepository struct {
	db *sqlx.DB
}

// NewReactionRepository 는 ReactionRepository를 생성한다.
func NewReactionRepository(db *sqlx.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// Create 는 새 반응을 저장한다.
//
// 같은 유저(session_id)가 같은 공고에 같은 타입의 반응을 중복으로 하면
// DB의 UNIQUE(announcement_id, session_id, type) 제약에 의해
// ON CONFLICT DO NOTHING으로 무시된다 (에러 없이 아무 일도 안 일어남).
//
// POST /api/announcements/:id/react API에서 호출한다.
func (r *ReactionRepository) Create(ctx context.Context, reaction *model.Reaction) error {
	query := `
		INSERT INTO reactions (announcement_id, type, session_id, subscriber_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (announcement_id, session_id, type) DO NOTHING
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		reaction.AnnouncementID, reaction.Type,
		reaction.SessionID, reaction.SubscriberID,
	).Scan(&reaction.ID, &reaction.CreatedAt)
}

// GetSummary 는 특정 공고에 대한 반응 유형별 집계(관심/찬성/반대 수)를 반환한다.
//
// GET /api/announcements/:id/reactions API에서 호출한다.
// 반응이 하나도 없으면 모든 카운트가 0인 ReactionSummary를 반환한다.
//
// 반환 예: {Interest: 847, Agree: 231, Disagree: 56}
func (r *ReactionRepository) GetSummary(ctx context.Context, announcementID string) (*model.ReactionSummary, error) {
	summary := &model.ReactionSummary{}
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'interest' THEN 1 ELSE 0 END), 0) AS interest,
			COALESCE(SUM(CASE WHEN type = 'agree' THEN 1 ELSE 0 END), 0) AS agree,
			COALESCE(SUM(CASE WHEN type = 'disagree' THEN 1 ELSE 0 END), 0) AS disagree
		FROM reactions
		WHERE announcement_id = $1`
	if err := r.db.GetContext(ctx, summary, query, announcementID); err != nil {
		return nil, fmt.Errorf("반응 집계 실패 (announcement_id=%s): %w", announcementID, err)
	}
	return summary, nil
}
