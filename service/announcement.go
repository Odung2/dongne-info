// Package service 는 비즈니스 로직 레이어를 담당한다.
//
// api(핸들러) 레이어에서 호출하며, repository 레이어를 통해 DB에 접근한다.
// 입력값 검증, 비즈니스 규칙 적용, 외부 API 호출 등을 이 레이어에서 처리한다.
// handler → service → repository 순서로 호출하며, 역방향 의존은 금지한다.
package service

import (
	"context"

	"dongne-info/model"
	"dongne-info/repository"
)

// AnnouncementService 는 공고 관련 비즈니스 로직을 담당한다.
//
// 현재는 repository를 직접 호출하는 패스스루(pass-through) 형태이지만,
// 향후 AI 요약 자동 연결, 관련 공고 매칭 등의 비즈니스 로직이 추가될 예정.
//
// 사용 예:
//
//	repo := repository.NewAnnouncementRepository(db)
//	svc := service.NewAnnouncementService(repo)
//	announcements, err := svc.List(ctx, filter)
type AnnouncementService struct {
	repo *repository.AnnouncementRepository
}

// NewAnnouncementService 는 AnnouncementService를 생성한다.
//
// 애플리케이션 시작 시 repository를 주입하여 한 번만 호출한다.
func NewAnnouncementService(repo *repository.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{repo: repo}
}

// GetByID 는 ID로 공고 1건을 조회한다.
//
// API의 GET /api/announcements/:id 핸들러에서 호출한다.
// 공고가 없으면 *model.NotFoundError를 반환한다.
func (s *AnnouncementService) GetByID(ctx context.Context, id string) (*model.Announcement, error) {
	return s.repo.FindByID(ctx, id)
}

// List 는 필터 조건에 맞는 공고 목록을 조회한다.
//
// API의 GET /api/announcements 핸들러에서 호출한다.
// 필터 조건이 비어있으면 전체 공고를 최신순으로 반환한다.
func (s *AnnouncementService) List(ctx context.Context, filter model.AnnouncementFilter) ([]model.Announcement, error) {
	return s.repo.FindByFilter(ctx, filter)
}
