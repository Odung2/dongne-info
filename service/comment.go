package service

import (
	"context"

	"dongne-info/model"
	"dongne-info/repository"
)

// CommentService 는 주민 의견(댓글) 관련 비즈니스 로직을 담당한다.
//
// 사용 예:
//
//	repo := repository.NewCommentRepository(db)
//	svc := service.NewCommentService(repo)
//	comment, err := svc.Create(ctx, announcementID, req)
type CommentService struct {
	repo *repository.CommentRepository
}

// NewCommentService 는 CommentService를 생성한다.
func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

// Create 는 공고에 대한 주민 의견을 저장한다.
//
// 로그인한 구독자만 작성 가능하다 (SubscriberID 필수).
// POST /api/announcements/:id/comments 핸들러에서 호출한다.
func (s *CommentService) Create(ctx context.Context, announcementID string, req model.CreateCommentRequest) (*model.Comment, error) {
	comment := &model.Comment{
		AnnouncementID: announcementID,
		SubscriberID:   req.SubscriberID,
		Body:           req.Body,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// ListByAnnouncement 는 특정 공고에 달린 의견 목록을 조회한다.
//
// 최신순으로 정렬하여 반환한다. 의견이 없으면 빈 슬라이스를 반환한다.
// GET /api/announcements/:id/comments 핸들러에서 호출한다.
func (s *CommentService) ListByAnnouncement(ctx context.Context, announcementID string, limit, offset int) ([]model.Comment, error) {
	return s.repo.FindByAnnouncementID(ctx, announcementID, limit, offset)
}
