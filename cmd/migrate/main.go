package main

import (
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL 환경변수가 설정되지 않았습니다")
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		slog.Error("DB 연결 실패", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("DB 연결 성공")

	migration, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		slog.Error("마이그레이션 파일 읽기 실패", "error", err)
		os.Exit(1)
	}

	if _, err := db.Exec(string(migration)); err != nil {
		slog.Error("마이그레이션 실행 실패", "error", err)
		os.Exit(1)
	}

	slog.Info("마이그레이션 완료")
}
