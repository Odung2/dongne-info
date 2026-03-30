package api

import (
	"net/http"
	"strconv"

	"dongne-info/model"
	"dongne-info/service"

	"github.com/gin-gonic/gin"
)

// AnnouncementHandler 는 공고 관련 HTTP 요청을 처리하는 핸들러.
//
// GET /api/announcements (목록), GET /api/announcements/:id (상세) 엔드포인트를 담당한다.
// SetupRouter에서 생성되어 라우터에 등록된다.
type AnnouncementHandler struct {
	svc *service.AnnouncementService
}

// NewAnnouncementHandler 는 AnnouncementHandler를 생성한다.
func NewAnnouncementHandler(svc *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc}
}

// List 는 공고 목록을 조회한다.
//
// GET /api/announcements?district=강남구&type=재개발&limit=20&offset=0
//
// 쿼리 파라미터:
//   - district: 자치구 필터 (선택)
//   - type: 공고 유형 필터 (선택)
//   - limit: 최대 조회 건수 (선택, 기본값 20)
//   - offset: 페이지네이션 시작점 (선택, 기본값 0)
//
// 응답: {"data": [...], "count": N}
func (h *AnnouncementHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := model.AnnouncementFilter{
		District: c.Query("district"),
		Type:     c.Query("type"),
		Limit:    limit,
		Offset:   offset,
	}

	announcements, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  announcements,
		"count": len(announcements),
	})
}

// Get 는 공고 1건을 상세 조회한다.
//
// GET /api/announcements/:id
//
// 공고가 존재하지 않으면 404를 반환한다.
//
// 응답: {"data": {...}}
func (h *AnnouncementHandler) Get(c *gin.Context) {
	id := c.Param("id")

	announcement, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": announcement})
}
