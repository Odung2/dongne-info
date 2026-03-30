package api

import (
	"net/http"

	"dongne-info/model"
	"dongne-info/service"

	"github.com/gin-gonic/gin"
)

// SubscriberHandler 는 구독 관련 HTTP 요청을 처리하는 핸들러.
//
// POST /api/subscribe (구독 등록), DELETE /api/subscribe/:id (구독 해지) 엔드포인트를 담당한다.
type SubscriberHandler struct {
	svc *service.SubscriberService
}

// NewSubscriberHandler 는 SubscriberHandler를 생성한다.
func NewSubscriberHandler(svc *service.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{svc: svc}
}

// Create 는 새 구독자를 등록한다.
//
// POST /api/subscribe
//
// 요청 바디:
//
//	{
//	  "contact": "user@example.com",
//	  "type": "email",           // "email" 또는 "kakao"
//	  "districts": ["강남구"]     // 구독할 자치구 목록
//	}
//
// 지원하지 않는 구를 입력하면 400 에러를 반환한다.
// 응답: {"data": {...}} (201 Created)
func (h *SubscriberHandler) Create(c *gin.Context) {
	var req model.CreateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청입니다"})
		return
	}

	subscriber, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": subscriber})
}

// Unsubscribe 는 구독을 해지한다.
//
// DELETE /api/subscribe/:id
//
// 실제 삭제하지 않고 active=FALSE로 변경한다.
// 해당 ID의 구독자가 없으면 404를 반환한다.
// 응답: {"message": "구독이 해지되었습니다"} (200 OK)
func (h *SubscriberHandler) Unsubscribe(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Unsubscribe(c.Request.Context(), id); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "구독이 해지되었습니다"})
}
