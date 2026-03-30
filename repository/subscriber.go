package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dongne-info/model"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SubscriberRepository 는 subscribers 테이블에 대한 CRUD를 담당한다.
//
// service.SubscriberService에서 사용하며,
// 구독 등록/해지/조회 기능을 제공한다.
type SubscriberRepository struct {
	db *sqlx.DB
}

// NewSubscriberRepository 는 SubscriberRepository를 생성한다.
//
// 애플리케이션 시작 시 DB 연결 후 한 번만 호출한다.
func NewSubscriberRepository(db *sqlx.DB) *SubscriberRepository {
	return &SubscriberRepository{db: db}
}

// FindByID 는 ID로 구독자 1명을 조회한다.
//
// 구독자가 존재하지 않으면 *model.NotFoundError를 반환한다.
func (r *SubscriberRepository) FindByID(ctx context.Context, id string) (*model.Subscriber, error) {
	var s model.Subscriber
	query := `SELECT * FROM subscribers WHERE id = $1`
	if err := r.db.GetContext(ctx, &s, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "구독자", ID: id}
		}
		return nil, fmt.Errorf("구독자 조회 실패 (id=%s): %w", id, err)
	}
	return &s, nil
}

// FindByDistrict 는 특정 구를 구독 중인 활성 구독자 목록을 조회한다.
//
// 알림 발송 시 사용한다. 새 공고가 "강남구"에 등록되면,
// districts 배열에 "강남구"가 포함된 활성(active=true) 구독자를 모두 조회한다.
//
// PostgreSQL의 ANY 연산자를 사용하여 배열 컬럼을 검색한다.
func (r *SubscriberRepository) FindByDistrict(ctx context.Context, district string) ([]model.Subscriber, error) {
	var subscribers []model.Subscriber
	query := `SELECT * FROM subscribers WHERE active = TRUE AND $1 = ANY(districts)`
	if err := r.db.SelectContext(ctx, &subscribers, query, district); err != nil {
		return nil, fmt.Errorf("구독자 목록 조회 실패 (district=%s): %w", district, err)
	}
	return subscribers, nil
}

// Create 는 새 구독자를 DB에 저장한다.
//
// POST /api/subscribe API에서 호출한다.
// 저장 성공 시 DB가 생성한 id와 created_at이 파라미터 구조체에 반영된다.
func (r *SubscriberRepository) Create(ctx context.Context, s *model.Subscriber) error {
	query := `
		INSERT INTO subscribers (contact, type, districts)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		s.Contact, s.Type, pq.Array(s.Districts),
	).Scan(&s.ID, &s.CreatedAt)
}

// Deactivate 는 구독자를 비활성화(구독 해지)한다.
//
// DELETE /api/subscribe/:id API에서 호출한다.
// 실제 삭제하지 않고 active=FALSE로 변경하여 데이터를 보존한다.
// 해당 ID의 구독자가 없으면 *model.NotFoundError를 반환한다.
func (r *SubscriberRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE subscribers SET active = FALSE WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("구독 해지 실패 (id=%s): %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &model.NotFoundError{Resource: "구독자", ID: id}
	}
	return nil
}
