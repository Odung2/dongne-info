package api

import (
	"net/http"
	"strconv"

	"dongne-info/model"
	"dongne-info/service"

	"github.com/gin-gonic/gin"
)

// CommentHandler 는 주민 의견(댓글) 관련 HTTP 요청을 처리하는 핸들러.
//
// POST /api/announcements/:id/comments (의견 작성),
// GET /api/announcements/:id/comments (목록 조회) 엔드포인트를 담당한다.
type CommentHandler struct {
	svc *service.CommentService
}

// NewCommentHandler 는 CommentHandler를 생성한다.
func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// Create 는 공고에 대한 주민 의견을 작성한다.
//
// POST /api/announcements/:id/comments
//
// 요청 바디:
//
//	{
//	  "body": "재건축 빨리 됐으면 좋겠네요",   // 의견 본문 (1~500자, 필수)
//	  "subscriber_id": "uuid-here"           // 작성자 구독자 ID (필수, 로그인 필요)
//	}
//
// 응답: {"data": {...}} (201 Created)
func (h *CommentHandler) Create(c *gin.Context) {
	announcementID := c.Param("id")

	var req model.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청입니다"})
		return
	}

	comment, err := h.svc.Create(c.Request.Context(), announcementID, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

// List 는 특정 공고에 달린 의견 목록을 조회한다.
//
// GET /api/announcements/:id/comments?limit=20&offset=0
//
// 쿼리 파라미터:
//   - limit: 최대 조회 건수 (선택, 기본값 20)
//   - offset: 페이지네이션 시작점 (선택, 기본값 0)
//
// 응답: {"data": [...], "count": N}
func (h *CommentHandler) List(c *gin.Context) {
	announcementID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	comments, err := h.svc.ListByAnnouncement(c.Request.Context(), announcementID, limit, offset)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  comments,
		"count": len(comments),
	})
}
