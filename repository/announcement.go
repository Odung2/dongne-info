// Package repository 는 DB 접근 레이어를 담당한다.
//
// sqlx를 사용하여 PostgreSQL(Supabase)에 직접 SQL 쿼리를 실행한다.
// ORM을 사용하지 않고 명시적인 SQL로 모든 쿼리를 작성하여
// 어떤 쿼리가 실행되는지 코드에서 바로 확인할 수 있다.
//
// 이 패키지의 모든 함수는 context.Context를 첫 번째 인자로 받아
// 타임아웃, 취소 등을 지원한다.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dongne-info/model"

	"github.com/jmoiron/sqlx"
)

// AnnouncementRepository 는 announcements 테이블에 대한 CRUD를 담당한다.
//
// service.AnnouncementService에서 사용하며,
// NewAnnouncementRepository(db)로 생성한다.
//
// 사용 예:
//
//	db, _ := sqlx.Connect("postgres", databaseURL)
//	repo := repository.NewAnnouncementRepository(db)
//	announcement, err := repo.FindByID(ctx, "some-uuid")
type AnnouncementRepository struct {
	db *sqlx.DB
}

// NewAnnouncementRepository 는 AnnouncementRepository를 생성한다.
//
// 애플리케이션 시작 시 DB 연결 후 한 번만 호출한다.
// 생성된 인스턴스는 service 레이어에 주입한다.
func NewAnnouncementRepository(db *sqlx.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// FindByID 는 ID로 공고 1건을 조회한다.
//
// 공고가 존재하지 않으면 *model.NotFoundError를 반환한다.
// API의 GET /api/announcements/:id 에서 사용한다.
func (r *AnnouncementRepository) FindByID(ctx context.Context, id string) (*model.Announcement, error) {
	var a model.Announcement
	query := `SELECT * FROM announcements WHERE id = $1`
	if err := r.db.GetContext(ctx, &a, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "공고", ID: id}
		}
		return nil, fmt.Errorf("공고 조회 실패 (id=%s): %w", id, err)
	}
	return &a, nil
}

// FindByFilter 는 필터 조건에 맞는 공고 목록을 조회한다.
//
// District, Type이 비어있으면 해당 조건을 무시하고 전체 조회한다.
// 결과는 created_at 내림차순(최신순)으로 정렬된다.
// API의 GET /api/announcements?district=강남구&type=재개발 에서 사용한다.
//
// 파라미터:
//   - filter.District: 필터할 자치구명 (빈 문자열이면 전체)
//   - filter.Type: 필터할 공고 유형 (빈 문자열이면 전체)
//   - filter.Limit: 최대 조회 건수 (0 이하면 기본값 20)
//   - filter.Offset: 페이지네이션 시작점
func (r *AnnouncementRepository) FindByFilter(ctx context.Context, filter model.AnnouncementFilter) ([]model.Announcement, error) {
	var announcements []model.Announcement

	query := `SELECT * FROM announcements WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.District != "" {
		query += fmt.Sprintf(" AND district = $%d", argIdx)
		args = append(args, filter.District)
		argIdx++
	}
	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)
	argIdx++

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	if err := r.db.SelectContext(ctx, &announcements, query, args...); err != nil {
		return nil, fmt.Errorf("공고 목록 조회 실패: %w", err)
	}
	return announcements, nil
}

// Create 는 새 공고를 DB에 저장한다.
//
// crawler 패키지에서 서울시 API 데이터를 파싱한 후 호출한다.
// 저장 성공 시 DB가 생성한 id와 created_at이 파라미터 구조체에 반영된다.
// source_id가 중복이면 DB UNIQUE 제약에 의해 에러가 발생한다.
func (r *AnnouncementRepository) Create(ctx context.Context, a *model.Announcement) error {
	query := `
		INSERT INTO announcements (district, type, action, title, location, summary, stage, related, raw_category, area_before, area_after, source_url, source_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		a.District, a.Type, a.Action, a.Title, a.Location,
		a.Summary, a.Stage, a.Related, a.RawCategory,
		a.AreaBefore, a.AreaAfter, a.SourceURL, a.SourceID,
	).Scan(&a.ID, &a.CreatedAt)
}

// ExistsBySourceID 는 해당 source_id의 공고가 이미 존재하는지 확인한다.
//
// 크롤러에서 중복 수집을 방지하기 위해 사용한다.
// 같은 공고를 여러 번 크롤링해도 한 번만 저장되도록 보장한다.
func (r *AnnouncementRepository) ExistsBySourceID(ctx context.Context, sourceID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM announcements WHERE source_id = $1)`
	if err := r.db.GetContext(ctx, &exists, query, sourceID); err != nil {
		return false, fmt.Errorf("중복 확인 실패 (source_id=%s): %w", sourceID, err)
	}
	return exists, nil
}

// SummaryData 는 AI 요약 결과를 담는 구조체.
//
// UpdateSummary에 전달하여 DB에 저장한다.
// 각 필드는 AI 프롬프트의 JSON 응답과 1:1 매핑.
type SummaryData struct {
	EasyTitle string // 쉬운 제목 (예: "공덕동 재개발 구역이 넓어졌어요"). 카드/알림에 사용.
	Summary   string // 쉬운 설명 2~3문장
	Stage     string // 진행 단계 (예: "2/7단계 - 정비구역지정")
	Related   string // 관련 사례 (없으면 빈 문자열)
	Impact    string // 나한테 어떤 의미? (예: "이 근처 거주자라면 확인해보세요")
	ActionTip string // 추천 액션 (예: "구청 도시계획과에 문의해보세요")
}

// UpdateSummary 는 공고의 AI 요약 정보를 업데이트한다.
//
// Claude API로 요약을 생성한 후 호출한다.
// 크롤링 시점에는 요약이 없고, 별도 프로세스에서 요약을 생성하여 업데이트하는 구조.
func (r *AnnouncementRepository) UpdateSummary(ctx context.Context, id string, data SummaryData) error {
	query := `UPDATE announcements SET easy_title = $1, summary = $2, stage = $3, related = $4, impact = $5, action_tip = $6 WHERE id = $7`
	result, err := r.db.ExecContext(ctx, query, data.EasyTitle, data.Summary, data.Stage, data.Related, data.Impact, data.ActionTip, id)
	if err != nil {
		return fmt.Errorf("요약 업데이트 실패 (id=%s): %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &model.NotFoundError{Resource: "공고", ID: id}
	}
	return nil
}

// FindUnsummarized 는 아직 AI 요약이 생성되지 않은 공고 목록을 조회한다.
//
// Claude API 요약 배치 작업에서 사용한다.
// summary 필드가 NULL인 공고를 최대 limit건 반환한다.
func (r *AnnouncementRepository) FindUnsummarized(ctx context.Context, limit int) ([]model.Announcement, error) {
	var announcements []model.Announcement
	if limit <= 0 {
		limit = 10
	}
	query := `SELECT * FROM announcements WHERE summary IS NULL ORDER BY created_at DESC LIMIT $1`
	if err := r.db.SelectContext(ctx, &announcements, query, limit); err != nil {
		return nil, fmt.Errorf("미요약 공고 조회 실패: %w", err)
	}
	return announcements, nil
}

// FindUnnotified 는 요약은 완료됐지만 아직 알림을 보내지 않은 공고를 조회한다.
//
// NotificationService에서 사용한다.
// summary가 NOT NULL이고 notified_at이 NULL인 공고를 반환한다.
func (r *AnnouncementRepository) FindUnnotified(ctx context.Context, limit int) ([]model.Announcement, error) {
	var announcements []model.Announcement
	if limit <= 0 {
		limit = 20
	}
	query := `SELECT * FROM announcements WHERE summary IS NOT NULL AND notified_at IS NULL ORDER BY created_at DESC LIMIT $1`
	if err := r.db.SelectContext(ctx, &announcements, query, limit); err != nil {
		return nil, fmt.Errorf("미발송 공고 조회 실패: %w", err)
	}
	return announcements, nil
}

// MarkNotified 는 공고의 알림 발송 시각을 현재 시각으로 업데이트한다.
//
// 알림 발송 완료 후 호출하여 중복 발송을 방지한다.
func (r *AnnouncementRepository) MarkNotified(ctx context.Context, id string) error {
	query := `UPDATE announcements SET notified_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("알림 발송 시각 업데이트 실패 (id=%s): %w", id, err)
	}
	return nil
}
