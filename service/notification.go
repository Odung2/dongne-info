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
// 해당 구를 구독한 활성 구독자에게 이메일을 발송한다.
//
// 사용 예:
//
//	emailSvc := service.NewEmailService(resendAPIKey)
//	notifSvc := service.NewNotificationService(subscriberRepo, announcementRepo, emailSvc)
//	count, err := notifSvc.NotifyNewAnnouncements(ctx)
type NotificationService struct {
	subscriberRepo   *repository.SubscriberRepository
	announcementRepo *repository.AnnouncementRepository
	emailSvc         *EmailService
}

// NewNotificationService 는 NotificationService를 생성한다.
//
// emailSvc가 nil이면 이메일 발송 없이 로그만 출력한다.
func NewNotificationService(
	subscriberRepo *repository.SubscriberRepository,
	announcementRepo *repository.AnnouncementRepository,
	emailSvc *EmailService,
) *NotificationService {
	return &NotificationService{
		subscriberRepo:   subscriberRepo,
		announcementRepo: announcementRepo,
		emailSvc:         emailSvc,
	}
}

// NotifyNewAnnouncements 는 미발송 공고에 대해 구독자에게 알림을 발송한다.
//
// summary가 있고 notified_at이 NULL인 공고를 찾아서,
// 해당 구의 구독자에게 이메일을 보내고 notified_at을 업데이트한다.
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
			_ = s.announcementRepo.MarkNotified(ctx, a.ID)
			notified++
			continue
		}

		for _, sub := range subscribers {
			if err := s.sendNotification(&a, &sub); err != nil {
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
//   - email: Resend API로 이메일 발송
//   - kakao: 미구현 (로그만 출력)
func (s *NotificationService) sendNotification(a *model.Announcement, sub *model.Subscriber) error {
	switch sub.Type {
	case "email":
		if s.emailSvc == nil {
			slog.Info("이메일 서비스 미설정, 로그만 출력",
				"to", sub.Contact, "title", a.Title)
			return nil
		}
		return s.emailSvc.SendAnnouncementEmail(sub.Contact, a)

	case "kakao":
		// TODO: 카카오 알림톡 연동
		slog.Info("카카오 알림톡 미구현, 로그만 출력",
			"to", sub.Contact, "title", a.Title)
		return nil

	default:
		return fmt.Errorf("지원하지 않는 구독 유형: %s", sub.Type)
	}
}

// FormatNotification 은 알림 메시지를 생성한다 (카카오톡용 텍스트 포맷).
func FormatNotification(a *model.Announcement) string {
	location := ptrOr(a.Location, "")
	summary := ptrOr(a.Summary, "")
	stage := ptrOr(a.Stage, "")

	stageText := ""
	if stage != "" {
		stageText = fmt.Sprintf("\n📊 %s", stage)
	}

	return fmt.Sprintf("📌 %s\n%s %s\n\n%s%s",
		a.Title, a.District, location, summary, stageText)
}
