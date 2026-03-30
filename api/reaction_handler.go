package api

import (
	"net/http"

	"dongne-info/model"
	"dongne-info/service"

	"github.com/gin-gonic/gin"
)

// ReactionHandler 는 공고 반응(관심/찬반) 관련 HTTP 요청을 처리하는 핸들러.
//
// POST /api/announcements/:id/react (반응 등록),
// GET /api/announcements/:id/reactions (집계 조회) 엔드포인트를 담당한다.
type ReactionHandler struct {
	svc *service.ReactionService
}

// NewReactionHandler 는 ReactionHandler를 생성한다.
func NewReactionHandler(svc *service.ReactionService) *ReactionHandler {
	return &ReactionHandler{svc: svc}
}

// Create 는 공고에 대한 반응(관심/찬성/반대)을 등록한다.
//
// POST /api/announcements/:id/react
//
// 요청 바디:
//
//	{
//	  "type": "interest",       // "interest", "agree", "disagree" 중 하나
//	  "session_id": "abc-123"   // 비로그인 유저 식별용 (선택)
//	}
//
// 같은 유저가 같은 공고에 같은 타입의 반응을 중복으로 하면 무시된다.
// 응답: {"data": {...}} (201 Created)
func (h *ReactionHandler) Create(c *gin.Context) {
	announcementID := c.Param("id")

	var req model.CreateReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청입니다"})
		return
	}

	reaction, err := h.svc.React(c.Request.Context(), announcementID, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": reaction})
}

// GetSummary 는 특정 공고에 대한 반응 유형별 집계를 반환한다.
//
// GET /api/announcements/:id/reactions
//
// 응답 예:
//
//	{"data": {"interest": 847, "agree": 231, "disagree": 56}}
func (h *ReactionHandler) GetSummary(c *gin.Context) {
	announcementID := c.Param("id")

	summary, err := h.svc.GetSummary(c.Request.Context(), announcementID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}
