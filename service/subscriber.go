package service

import (
	"context"
	"fmt"

	"dongne-info/model"
	"dongne-info/repository"
)

// validDistricts 는 현재 서비스가 지원하는 자치구 목록.
// 구독 등록 시 이 목록에 없는 구를 입력하면 ValidationError를 반환한다.
// 서비스 확장 시 여기에 구를 추가하면 된다.
var validDistricts = map[string]bool{
	"강남구": true,
	"마포구": true,
}

// SubscriberService 는 구독 관련 비즈니스 로직을 담당한다.
//
// 구독 등록 시 지원하는 구인지 검증하고, 구독 해지를 처리한다.
//
// 사용 예:
//
//	repo := repository.NewSubscriberRepository(db)
//	svc := service.NewSubscriberService(repo)
//	subscriber, err := svc.Create(ctx, req)
type SubscriberService struct {
	repo *repository.SubscriberRepository
}

// NewSubscriberService 는 SubscriberService를 생성한다.
func NewSubscriberService(repo *repository.SubscriberRepository) *SubscriberService {
	return &SubscriberService{repo: repo}
}

// Create 는 새 구독자를 등록한다.
//
// 등록 전에 요청된 자치구가 서비스에서 지원하는 구인지 검증한다.
// 지원하지 않는 구가 포함되면 *model.ValidationError를 반환한다.
//
// POST /api/subscribe 핸들러에서 호출한다.
func (s *SubscriberService) Create(ctx context.Context, req model.CreateSubscriberRequest) (*model.Subscriber, error) {
	// 지원하는 구인지 검증
	for _, d := range req.Districts {
		if !validDistricts[d] {
			return nil, &model.ValidationError{
				Field:   "districts",
				Message: fmt.Sprintf("지원하지 않는 구입니다: %s", d),
			}
		}
	}

	subscriber := &model.Subscriber{
		Contact:   req.Contact,
		Type:      req.Type,
		Districts: req.Districts,
		Active:    true,
	}

	if err := s.repo.Create(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("구독자 생성 실패: %w", err)
	}
	return subscriber, nil
}

// Unsubscribe 는 구독자를 비활성화(구독 해지)한다.
//
// 실제 삭제하지 않고 active=FALSE로 변경한다.
// DELETE /api/subscribe/:id 핸들러에서 호출한다.
func (s *SubscriberService) Unsubscribe(ctx context.Context, id string) error {
	return s.repo.Deactivate(ctx, id)
}
