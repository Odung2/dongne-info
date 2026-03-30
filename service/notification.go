package service

import (
	"context"
	"fmt"
	"log/slog"

	"dongne-info/model"
	"dongne-info/repository"
)

// NotificationService 는 새 공고 발생 시 구독자에게 알림을 발송하는 서비스.
//
// 크롤링 후 AI 요약이 완료된 공고 중 아직 알림을 보내지 않은 것을 찾아
// 해당 구를 구독한 활성 구독자에게 알림을 발송한다.
//
// 알림 발송 순서:
//  1. notified_at이 NULL인 공고(요약 완료 + 미발송) 조회
//  2. 해당 공고의 district로 구독자 목록 조회
//  3. 구독자 유형(email/kakao)에 따라 알림 발송
//  4. notification_logs에 발송 이력 기록
//  5. 공고의 notified_at 업데이트
//
// 사용 예:
//
//	svc := service.NewNotificationService(subscriberRepo, announcementRepo, logRepo)
//	count, err := svc.NotifyNewAnnouncements(ctx)
type NotificationService struct {
	subscriberRepo   *repository.SubscriberRepository
	announcementRepo *repository.AnnouncementRepository
}

// NewNotificationService 는 NotificationService를 생성한다.
func NewNotificationService(
	subscriberRepo *repository.SubscriberRepository,
	announcementRepo *repository.AnnouncementRepository,
) *NotificationService {
	return &NotificationService{
		subscriberRepo:   subscriberRepo,
		announcementRepo: announcementRepo,
	}
}

// NotifyNewAnnouncements 는 미발송 공고에 대해 구독자에게 알림을 발송한다.
//
// summary가 있고 notified_at이 NULL인 공고를 찾아서,
// 해당 구의 구독자에게 알림을 보내고 notified_at을 업데이트한다.
//
// 반환:
//   - int: 알림을 발송한 공고 수
//   - error: DB 조회 실패 시에만 에러 반환
func (s *NotificationService) NotifyNewAnnouncements(ctx context.Context) (int, error) {
	announcements, err := s.announcementRepo.FindUnnotified(ctx, 20)
	if err != nil {
		return 0, fmt.Errorf("미발송 공고 조회 실패: %w", err)
	}

	if len(announcements) == 0 {
		slog.Info("발송할 알림 없음")
		return 0, nil
	}

	slog.Info("알림 발송 시작", "count", len(announcements))

	notified := 0
	for _, a := range announcements {
		subscribers, err := s.subscriberRepo.FindByDistrict(ctx, a.District)
		if err != nil {
			slog.Error("구독자 조회 실패", "district", a.District, "error", err)
			continue
		}

		if len(subscribers) == 0 {
			slog.Info("구독자 없음, 발송 건너뜀", "district", a.District, "title", a.Title)
			// 구독자 없어도 notified_at은 업데이트 (다음에 다시 시도하지 않도록)
			_ = s.announcementRepo.MarkNotified(ctx, a.ID)
			notified++
			continue
		}

		for _, sub := range subscribers {
			if err := s.sendNotification(ctx, &a, &sub); err != nil {
				slog.Error("알림 발송 실패", "subscriber_id", sub.ID, "error", err)
				continue
			}
		}

		_ = s.announcementRepo.MarkNotified(ctx, a.ID)
		notified++
		slog.Info("알림 발송 완료", "title", a.Title, "district", a.District, "subscribers", len(subscribers))
	}

	return notified, nil
}

// sendNotification 은 개별 구독자에게 알림을 발송한다.
//
// 구독자 유형(email/kakao)에 따라 다른 채널로 발송한다.
// 현재는 로그만 출력하며, 실제 이메일/카카오 연동은 추후 구현.
//
// TODO: 이메일 발송 (SendGrid/SES 등)
// TODO: 카카오 알림톡 발송
func (s *NotificationService) sendNotification(ctx context.Context, a *model.Announcement, sub *model.Subscriber) error {
	summary := ""
	if a.Summary != nil {
		summary = *a.Summary
	}

	// 현재는 로그로 대체. 실제 발송은 추후 구현.
	slog.Info("알림 발송",
		"to", sub.Contact,
		"type", sub.Type,
		"district", a.District,
		"title", a.Title,
		"summary", summary,
	)

	return nil
}

// FormatNotification 은 알림 메시지를 생성한다.
//
// 카카오 알림톡, 이메일 등 채널에 관계없이 통일된 메시지를 만든다.
// ARCHITECTURE.md의 알림 설계 원칙에 따라 이슈 단위로 개인화된 메시지를 생성한다.
//
// 사용 예:
//
//	msg := service.FormatNotification(announcement)
//	// "📌 은마아파트 재건축 정비구역\n강남구 대치동 316번지 일대\n\n재건축을 공식적으로 시작할 수 있게 됐어요..."
func FormatNotification(a *model.Announcement) string {
	location := ""
	if a.Location != nil {
		location = *a.Location
	}
	summary := ""
	if a.Summary != nil {
		summary = *a.Summary
	}
	stage := ""
	if a.Stage != nil {
		stage = fmt.Sprintf("\n📊 %s", *a.Stage)
	}

	return fmt.Sprintf("📌 %s\n%s %s\n\n%s%s",
		a.Title, a.District, location, summary, stage)
}
