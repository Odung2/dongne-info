package api

import (
	"errors"
	"net/http"

	"dongne-info/model"

	"github.com/gin-gonic/gin"
)

// HandleError 는 커스텀 에러 타입에 따라 적절한 HTTP 상태 코드와 에러 메시지를 응답한다.
//
// 모든 핸들러에서 에러 발생 시 이 함수를 호출하여 일관된 에러 응답을 보장한다.
// 에러 타입과 HTTP 상태 코드 매핑:
//   - *model.NotFoundError    → 404 Not Found (리소스 없음)
//   - *model.ValidationError  → 400 Bad Request (입력값 오류)
//   - *model.ExternalAPIError → 502 Bad Gateway (외부 API 장애)
//   - 그 외 모든 에러         → 500 Internal Server Error
//
// 사용 예:
//
//	announcement, err := h.svc.GetByID(ctx, id)
//	if err != nil {
//	    HandleError(c, err)
//	    return
//	}
//
// 주의: ExternalAPIError의 경우 원본 에러 메시지를 노출하지 않고
// 일반적인 메시지("외부 서비스 오류가 발생했습니다")를 반환한다.
// 원본 에러는 서버 로그에만 기록해야 한다.
func HandleError(c *gin.Context, err error) {
	var notFound *model.NotFoundError
	if errors.As(err, &notFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var validation *model.ValidationError
	if errors.As(err, &validation) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var externalAPI *model.ExternalAPIError
	if errors.As(err, &externalAPI) {
		c.JSON(http.StatusBadGateway, gin.H{"error": "외부 서비스 오류가 발생했습니다"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "서버 오류가 발생했습니다"})
}
