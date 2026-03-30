package model_test

import (
	"errors"
	"fmt"
	"testing"

	"dongne-info/model"
)

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      *model.NotFoundError
		wantMsg  string
	}{
		{
			name:    "공고 조회 실패",
			err:     &model.NotFoundError{Resource: "공고", ID: "abc-123"},
			wantMsg: "공고을(를) 찾을 수 없습니다 (id=abc-123)",
		},
		{
			name:    "구독자 조회 실패",
			err:     &model.NotFoundError{Resource: "구독자", ID: "xyz-456"},
			wantMsg: "구독자을(를) 찾을 수 없습니다 (id=xyz-456)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &model.ValidationError{Field: "district", Message: "지원하지 않는 구입니다"}
	want := "입력값 오류: district - 지원하지 않는 구입니다"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestExternalAPIError_Unwrap(t *testing.T) {
	original := fmt.Errorf("connection timeout")
	err := &model.ExternalAPIError{Service: "서울시 API", Err: original}

	if !errors.Is(err, original) {
		t.Error("Unwrap()이 원본 에러를 반환해야 합니다")
	}

	want := "외부 API 오류 (서울시 API): connection timeout"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorTypeAssertion(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		isNotFound bool
		isValidation bool
		isExternal   bool
	}{
		{
			name:       "NotFoundError 타입 확인",
			err:        &model.NotFoundError{Resource: "공고", ID: "1"},
			isNotFound: true,
		},
		{
			name:         "ValidationError 타입 확인",
			err:          &model.ValidationError{Field: "a", Message: "b"},
			isValidation: true,
		},
		{
			name:       "ExternalAPIError 타입 확인",
			err:        &model.ExternalAPIError{Service: "test", Err: fmt.Errorf("fail")},
			isExternal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var notFound *model.NotFoundError
			var validation *model.ValidationError
			var external *model.ExternalAPIError

			if errors.As(tt.err, &notFound) != tt.isNotFound {
				t.Errorf("NotFoundError 매칭 실패")
			}
			if errors.As(tt.err, &validation) != tt.isValidation {
				t.Errorf("ValidationError 매칭 실패")
			}
			if errors.As(tt.err, &external) != tt.isExternal {
				t.Errorf("ExternalAPIError 매칭 실패")
			}
		})
	}
}
