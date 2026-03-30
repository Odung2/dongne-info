// Package config 는 애플리케이션의 환경변수 설정을 관리한다.
//
// .env 파일 또는 시스템 환경변수에서 값을 읽어 Config 구조체로 파싱한다.
// 애플리케이션 시작 시 config.Load()를 호출하여 설정을 로드하고,
// 필수 값이 없으면 에러를 반환한다.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config 는 애플리케이션 전체에서 사용하는 설정 값을 담는 구조체.
//
// 서버 시작 시 config.Load()로 생성하고, 각 레이어(서버, 크롤러, 서비스)에
// 필요한 값을 전달할 때 사용한다.
//
// 필수 환경변수가 없으면 Load() 시점에 에러가 발생한다.
// .env.example 파일에 전체 환경변수 목록이 정의되어 있다.
//
// 필드:
//   - Port: API 서버 포트 (기본값 "8080")
//   - DatabaseURL: Supabase PostgreSQL 연결 문자열 (필수)
//   - SeoulAPIKey: 서울시 열린데이터광장 API 인증키 (필수)
//   - AnthropicKey: Claude API 키. 공고 AI 요약에 사용 (선택, 없으면 요약 건너뜀)
//   - KakaoAPIKey: 카카오 알림톡 API 키 (선택, 없으면 알림 건너뜀)
//   - ResendAPIKey: Resend 이메일 발송 API 키 (선택, 없으면 이메일 알림 건너뜀)
//   - Environment: 실행 환경. "development" 또는 "production" (기본값 "development")
type Config struct {
	Port         string `env:"PORT" envDefault:"8080"`
	DatabaseURL  string `env:"DATABASE_URL,required"`
	SeoulAPIKey  string `env:"SEOUL_API_KEY,required"`
	AnthropicKey string `env:"ANTHROPIC_API_KEY"`
	KakaoAPIKey  string `env:"KAKAO_API_KEY"`
	ResendAPIKey string `env:"RESEND_API_KEY"`
	Environment  string `env:"ENVIRONMENT" envDefault:"development"`
}

// Load 는 환경변수에서 설정을 읽어 Config 구조체를 반환한다.
//
// 애플리케이션의 가장 첫 단계에서 호출해야 한다.
// DATABASE_URL, SEOUL_API_KEY 등 필수 환경변수가 없으면 에러를 반환한다.
//
// 사용 예:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	return cfg, nil
}
