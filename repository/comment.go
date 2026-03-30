package repository

import (
	"context"
	"fmt"

	"dongne-info/model"

	"github.com/jmoiron/sqlx"
)

// CommentRepository 는 comments 테이블에 대한 CRUD를 담당한다.
//
// 주민의 한줄 의견을 저장하고 조회한다.
// service.CommentService에서 사용한다.
type CommentRepository struct {
	db *sqlx.DB
}

// NewCommentRepository 는 CommentRepository를 생성한다.
func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create 는 새 의견을 DB에 저장한다.
//
// POST /api/announcements/:id/comments API에서 호출한다.
// 로그인한 구독자만 작성 가능하므로 subscriber_id가 필수이다.
// 저장 성공 시 DB가 생성한 id와 created_at이 파라미터 구조체에 반영된다.
func (r *CommentRepository) Create(ctx context.Context, c *model.Comment) error {
	query := `
		INSERT INTO comments (announcement_id, subscriber_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		c.AnnouncementID, c.SubscriberID, c.Body,
	).Scan(&c.ID, &c.CreatedAt)
}

// FindByAnnouncementID 는 특정 공고에 달린 의견 목록을 조회한다.
//
// GET /api/announcements/:id/comments API에서 호출한다.
// 최신순(created_at DESC)으로 정렬하여 반환한다.
// 의견이 없으면 빈 슬라이스를 반환한다 (에러 아님).
//
// 파라미터:
//   - announcementID: 조회할 공고 ID
//   - limit: 최대 조회 건수 (0 이하면 기본값 20)
//   - offset: 페이지네이션 시작점
func (r *CommentRepository) FindByAnnouncementID(ctx context.Context, announcementID string, limit, offset int) ([]model.Comment, error) {
	var comments []model.Comment
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT * FROM comments
		WHERE announcement_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &comments, query, announcementID, limit, offset); err != nil {
		return nil, fmt.Errorf("의견 목록 조회 실패 (announcement_id=%s): %w", announcementID, err)
	}
	return comments, nil
}
