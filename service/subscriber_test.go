package service_test

import (
	"context"
	"errors"
	"testing"

	"dongne-info/model"
	"dongne-info/service"
)

func TestSubscriberService_Create_InvalidDistrict(t *testing.T) {
	// repo 없이 validation 로직만 테스트 (DB 호출 전에 실패해야 함)
	svc := service.NewSubscriberService(nil)

	tests := []struct {
		name      string
		districts []string
		wantErr   bool
	}{
		{
			name:      "지원하지 않는 구",
			districts: []string{"종로구"},
			wantErr:   true,
		},
		{
			name:      "유효 + 무효 혼합",
			districts: []string{"강남구", "서초구"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := model.CreateSubscriberRequest{
				Contact:   "test@test.com",
				Type:      "email",
				Districts: tt.districts,
			}

			_, err := svc.Create(context.Background(), req)

			if tt.wantErr {
				if err == nil {
					t.Error("에러가 발생해야 합니다")
				}
				var validationErr *model.ValidationError
				if !errors.As(err, &validationErr) {
					t.Error("ValidationError 타입이어야 합니다")
				}
			}
			// wantErr=false인 경우 repo가 nil이라 panic 발생 → validation만 테스트
		})
	}
}
