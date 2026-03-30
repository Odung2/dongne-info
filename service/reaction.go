package service

import (
	"context"

	"dongne-info/model"
	"dongne-info/repository"
)

// ReactionService 는 공고 반응(관심/찬반) 관련 비즈니스 로직을 담당한다.
//
// 사용 예:
//
//	repo := repository.NewReactionRepository(db)
//	svc := service.NewReactionService(repo)
//	summary, err := svc.GetSummary(ctx, announcementID)
type ReactionService struct {
	repo *repository.ReactionRepository
}

// NewReactionService 는 ReactionService를 생성한다.
func NewReactionService(repo *repository.ReactionRepository) *ReactionService {
	return &ReactionService{repo: repo}
}

// React 는 공고에 대한 반응(관심/찬성/반대)을 등록한다.
//
// 같은 유저가 같은 공고에 같은 타입의 반응을 중복으로 하면
// DB UNIQUE 제약에 의해 무시된다 (에러 없이 아무 일도 안 일어남).
//
// POST /api/announcements/:id/react 핸들러에서 호출한다.
func (s *ReactionService) React(ctx context.Context, announcementID string, req model.CreateReactionRequest) (*model.Reaction, error) {
	reaction := &model.Reaction{
		AnnouncementID: announcementID,
		Type:           req.Type,
		SessionID:      req.SessionID,
	}

	if err := s.repo.Create(ctx, reaction); err != nil {
		return nil, err
	}
	return reaction, nil
}

// GetSummary 는 특정 공고에 대한 반응 유형별 집계를 반환한다.
//
// 관심/찬성/반대 각각의 수를 카운트하여 ReactionSummary로 반환한다.
// GET /api/announcements/:id/reactions 핸들러에서 호출한다.
func (s *ReactionService) GetSummary(ctx context.Context, announcementID string) (*model.ReactionSummary, error) {
	return s.repo.GetSummary(ctx, announcementID)
}
