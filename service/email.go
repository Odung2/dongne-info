package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"dongne-info/model"
)

// EmailService 는 Resend API를 사용하여 이메일을 발송하는 서비스.
//
// 새 공고가 수집되면 NotificationService에서 이 서비스를 호출하여
// 구독자에게 공고 요약 이메일을 보낸다.
//
// Resend 무료 플랜: 일 100건, 월 3,000건.
// 도메인 연결 전에는 발신자가 onboarding@resend.dev로 표시된다.
//
// 사용 예:
//
//	emailSvc := service.NewEmailService(resendAPIKey)
//	err := emailSvc.SendAnnouncementEmail("user@example.com", announcement)
type EmailService struct {
	apiKey string
	from   string // 발신자 이메일 (도메인 연결 전: onboarding@resend.dev)
}

// NewEmailService 는 EmailService를 생성한다.
// apiKey가 빈 문자열이면 nil을 반환한다 (이메일 기능 비활성화).
func NewEmailService(apiKey string) *EmailService {
	if apiKey == "" {
		return nil
	}
	return &EmailService{
		apiKey: apiKey,
		from:   "onboarding@resend.dev", // 도메인 연결 후 "동깨 <noreply@dongkkae.kr>"로 변경
	}
}

// resendRequest 는 Resend API 요청 바디.
type resendRequest struct {
	From    string `json:"from"`
	To      []string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// SendAnnouncementEmail 은 공고 요약 이메일을 발송한다.
//
// 이메일 내용:
//   - 제목: "[동깨] 사업명"
//   - 본문: AI 요약 + 진행 단계 + 나한테 어떤 의미 + 추천 액션
//
// 파라미터:
//   - to: 수신자 이메일 주소
//   - a: 공고 데이터 (요약 포함)
func (s *EmailService) SendAnnouncementEmail(to string, a *model.Announcement) error {
	subject := fmt.Sprintf("[동깨] %s", a.Title)
	html := buildEmailHTML(a)

	reqBody := resendRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("이메일 요청 생성 실패: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("HTTP 요청 생성 실패: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &model.ExternalAPIError{Service: "Resend", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &model.ExternalAPIError{
			Service: "Resend",
			Err:     fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	slog.Info("이메일 발송 성공", "to", to, "subject", subject)
	return nil
}

// buildEmailHTML 은 공고 데이터로 이메일 HTML을 생성한다.
//
// 심플한 HTML 레이아웃. 모바일에서도 잘 보이도록 인라인 스타일 사용.
func buildEmailHTML(a *model.Announcement) string {
	summary := ptrOr(a.Summary, "")
	stage := ptrOr(a.Stage, "")
	impact := ptrOr(a.Impact, "")
	actionTip := ptrOr(a.ActionTip, "")
	location := ptrOr(a.Location, "")

	stageHTML := ""
	if stage != "" {
		stageHTML = fmt.Sprintf(`<p style="margin:8px 0;padding:8px 12px;background:#EBF5FF;border-radius:8px;font-size:14px;color:#1D4ED8;">📊 %s</p>`, stage)
	}

	impactHTML := ""
	if impact != "" {
		impactHTML = fmt.Sprintf(`<div style="margin:12px 0;padding:12px;background:#EFF6FF;border-radius:8px;">
			<p style="margin:0 0 4px;font-size:12px;font-weight:600;color:#1E40AF;">나한테 어떤 의미?</p>
			<p style="margin:0;font-size:14px;color:#1E40AF;">%s</p>
		</div>`, impact)
	}

	actionHTML := ""
	if actionTip != "" {
		actionHTML = fmt.Sprintf(`<div style="margin:12px 0;padding:12px;background:#EFF6FF;border-radius:8px;">
			<p style="margin:0 0 4px;font-size:12px;font-weight:600;color:#1E40AF;">이런 걸 해볼 수 있어요</p>
			<p style="margin:0;font-size:14px;color:#1E40AF;">%s</p>
		</div>`, actionTip)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#F3F4F6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
<div style="max-width:480px;margin:0 auto;padding:20px;">
	<div style="background:white;border-radius:12px;padding:24px;box-shadow:0 1px 3px rgba(0,0,0,0.1);">
		<p style="margin:0 0 4px;font-size:12px;color:#6B7280;">%s · %s</p>
		<h1 style="margin:0 0 8px;font-size:18px;color:#111827;line-height:1.4;">%s</h1>
		%s
		<hr style="border:none;border-top:1px solid #E5E7EB;margin:16px 0;">
		<p style="margin:0;font-size:15px;color:#374151;line-height:1.7;">%s</p>
		%s
		%s
		%s
		<hr style="border:none;border-top:1px solid #E5E7EB;margin:16px 0;">
		<p style="margin:0;font-size:12px;color:#9CA3AF;text-align:center;">
			동깨 — 우리 동네 소식, 쉽게 알려줘요
		</p>
	</div>
</div>
</body>
</html>`, a.District, a.Type, a.Title,
		func() string {
			if location != "" {
				return fmt.Sprintf(`<p style="margin:0 0 12px;font-size:13px;color:#6B7280;">📍 %s</p>`, location)
			}
			return ""
		}(),
		summary, stageHTML, impactHTML, actionHTML)
}
