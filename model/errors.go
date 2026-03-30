// Package model 은 애플리케이션 전체에서 사용하는 데이터 구조체와 에러 타입을 정의한다.
//
// 이 패키지의 구조체는 DB 모델, API 요청/응답, 에러 타입으로 나뉜다.
// 다른 모든 패키지(repository, service, api, crawler)에서 import하여 사용한다.
package model

import "fmt"

// NotFoundError 는 요청한 리소스가 DB에 존재하지 않을 때 반환하는 에러.
//
// repository 레이어에서 sql.ErrNoRows를 감지했을 때 이 에러로 변환하여 반환한다.
// api 레이어의 HandleError에서 이 타입을 감지하면 HTTP 404를 응답한다.
//
// 사용 예:
//
//	return nil, &model.NotFoundError{Resource: "공고", ID: id}
//
// 필드:
//   - Resource: 리소스 종류 ("공고", "구독자" 등)
//   - ID: 조회 시도한 ID 값
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s을(를) 찾을 수 없습니다 (id=%s)", e.Resource, e.ID)
}

// ValidationError 는 사용자 입력값이 유효하지 않을 때 반환하는 에러.
//
// service 레이어에서 비즈니스 규칙 검증 실패 시 이 에러로 반환한다.
// 예: 지원하지 않는 구(district) 입력, 필수 필드 누락 등.
// api 레이어의 HandleError에서 이 타입을 감지하면 HTTP 400을 응답한다.
//
// 사용 예:
//
//	return nil, &model.ValidationError{Field: "districts", Message: "지원하지 않는 구입니다: 종로구"}
//
// 필드:
//   - Field: 검증 실패한 필드명
//   - Message: 사용자에게 보여줄 에러 메시지
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("입력값 오류: %s - %s", e.Field, e.Message)
}

// ExternalAPIError 는 외부 API(서울시 API, Claude API 등) 호출 실패 시 반환하는 에러.
//
// crawler, service 레이어에서 외부 API 요청이 실패했을 때 이 에러로 래핑하여 반환한다.
// Unwrap()을 구현하여 원본 에러를 errors.Is()로 확인할 수 있다.
// api 레이어의 HandleError에서 이 타입을 감지하면 HTTP 502를 응답한다.
//
// 사용 예:
//
//	return nil, &model.ExternalAPIError{Service: "서울시 upisRebuild API", Err: err}
//
// 필드:
//   - Service: 실패한 외부 서비스 이름 (로그/디버깅용)
//   - Err: 원본 에러 (Unwrap으로 접근 가능)
type ExternalAPIError struct {
	Service string
	Err     error
}

func (e *ExternalAPIError) Error() string {
	return fmt.Sprintf("외부 API 오류 (%s): %v", e.Service, e.Err)
}

// Unwrap 은 원본 에러를 반환한다. errors.Is(), errors.As()에서 사용된다.
func (e *ExternalAPIError) Unwrap() error {
	return e.Err
}
