package main

import (
	"log/slog"
	"os"

	"dongne-info/api"
	"dongne-info/config"
	"dongne-info/repository"
	"dongne-info/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 설정 로드
	cfg, err := config.Load()
	if err != nil {
		slog.Error("설정 로드 실패", "error", err)
		os.Exit(1)
	}

	// DB 연결
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("DB 연결 실패", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("DB 연결 성공")

	// Repository
	announcementRepo := repository.NewAnnouncementRepository(db)
	subscriberRepo := repository.NewSubscriberRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Service
	announcementSvc := service.NewAnnouncementService(announcementRepo)
	subscriberSvc := service.NewSubscriberService(subscriberRepo)
	reactionSvc := service.NewReactionService(reactionRepo)
	commentSvc := service.NewCommentService(commentRepo)

	// Router
	router := api.SetupRouter(
		announcementSvc,
		subscriberSvc,
		reactionSvc,
		commentSvc,
	)

	slog.Info("서버 시작", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("서버 실행 실패", "error", err)
		os.Exit(1)
	}
}
