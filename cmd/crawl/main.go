// cmd/crawl 은 크롤링 + AI 요약을 실행하는 CLI 진입점.
//
// 실행 순서:
//  1. 강남구/마포구 재개발 데이터를 서울시 API에서 수집
//  2. 새 공고를 DB에 저장
//  3. 아직 요약되지 않은 공고에 대해 Claude API로 AI 요약 생성
//
// 사용:
//
//	export DATABASE_URL=... SEOUL_API_KEY=... ANTHROPIC_API_KEY=...
//	go run cmd/crawl/main.go
//
// 스케줄러(cron)에서 매일 오전 9시에 자동 실행하도록 설정한다.
package main

import (
	"context"
	"log/slog"
	"os"

	"dongne-info/config"
	"dongne-info/crawler"
	"dongne-info/repository"
	"dongne-info/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("설정 로드 실패", "error", err)
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("DB 연결 실패", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewAnnouncementRepository(db)

	// 1단계: 크롤링
	rebuildCrawler := crawler.NewRebuildCrawler(cfg.SeoulAPIKey, repo)

	ctx := context.Background()
	districts := []string{"강남구", "마포구"}
	totalNew := 0

	for _, district := range districts {
		slog.Info("크롤링 시작", "district", district)
		count, err := rebuildCrawler.Crawl(ctx, district)
		if err != nil {
			slog.Error("크롤링 실패", "district", district, "error", err)
			continue
		}
		totalNew += count
		slog.Info("크롤링 완료", "district", district, "new_count", count)
	}

	// 1-2단계: 도시계획 결정고시 크롤링
	cityplanCrawler := crawler.NewCityplanCrawler(cfg.SeoulAPIKey, repo)
	slog.Info("도시계획 크롤링 시작")
	cpCount, err := cityplanCrawler.Crawl(ctx, districts)
	if err != nil {
		slog.Error("도시계획 크롤링 실패", "error", err)
	} else {
		totalNew += cpCount
		slog.Info("도시계획 크롤링 완료", "new_count", cpCount)
	}

	slog.Info("전체 크롤링 완료", "total_new", totalNew)

	// 2단계: AI 요약 (ANTHROPIC_API_KEY가 있을 때만)
	summarySvc := service.NewSummaryService(cfg.AnthropicKey, repo)
	if summarySvc == nil {
		slog.Warn("ANTHROPIC_API_KEY 미설정, AI 요약 건너뜀")
		return
	}

	summarized, err := summarySvc.SummarizeUnsummarized(ctx, 20)
	if err != nil {
		slog.Error("AI 요약 실패", "error", err)
		return
	}
	slog.Info("AI 요약 완료", "summarized_count", summarized)

	// 3단계: 알림 발송
	subscriberRepo := repository.NewSubscriberRepository(db)
	notificationSvc := service.NewNotificationService(subscriberRepo, repo)
	notifiedCount, err := notificationSvc.NotifyNewAnnouncements(ctx)
	if err != nil {
		slog.Error("알림 발송 실패", "error", err)
		return
	}
	slog.Info("알림 발송 완료", "notified_count", notifiedCount)
}
