// Package api 는 HTTP 핸들러와 라우팅을 담당한다.
//
// Gin 프레임워크를 사용하며, 모든 엔드포인트는 /api 접두사 아래에 정의된다.
// 각 핸들러는 service 레이어를 호출하고, 에러 발생 시 HandleError로 통일된 응답을 반환한다.
//
// 엔드포인트 목록:
//   - GET    /api/announcements           공고 목록 조회
//   - GET    /api/announcements/:id        공고 상세 조회
//   - POST   /api/subscribe               구독 등록
//   - DELETE /api/subscribe/:id            구독 해지
//   - POST   /api/announcements/:id/react  반응(관심/찬반) 등록
//   - GET    /api/announcements/:id/reactions  반응 집계 조회
//   - POST   /api/announcements/:id/comments   의견 작성
//   - GET    /api/announcements/:id/comments   의견 목록 조회
//   - GET    /api/districts               지원 구 목록 조회
package api

import (
	"dongne-info/api/middleware"
	"dongne-info/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter 는 Gin 라우터를 생성하고 모든 엔드포인트를 등록한다.
//
// cmd/server/main.go에서 서비스 인스턴스들을 주입받아 호출한다.
// 반환된 *gin.Engine의 Run() 메서드로 서버를 시작한다.
//
// 사용 예:
//
//	router := api.SetupRouter(announcementSvc, subscriberSvc, reactionSvc, commentSvc)
//	router.Run(":8080")
func SetupRouter(
	announcementSvc *service.AnnouncementService,
	subscriberSvc *service.SubscriberService,
	reactionSvc *service.ReactionService,
	commentSvc *service.CommentService,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	announcementHandler := NewAnnouncementHandler(announcementSvc)
	subscriberHandler := NewSubscriberHandler(subscriberSvc)
	reactionHandler := NewReactionHandler(reactionSvc)
	commentHandler := NewCommentHandler(commentSvc)

	api := r.Group("/api")
	{
		// 공고
		api.GET("/announcements", announcementHandler.List)
		api.GET("/announcements/:id", announcementHandler.Get)

		// 구독
		api.POST("/subscribe", subscriberHandler.Create)
		api.DELETE("/subscribe/:id", subscriberHandler.Unsubscribe)

		// 반응
		api.POST("/announcements/:id/react", reactionHandler.Create)
		api.GET("/announcements/:id/reactions", reactionHandler.GetSummary)

		// 의견
		api.POST("/announcements/:id/comments", commentHandler.Create)
		api.GET("/announcements/:id/comments", commentHandler.List)

		// 지원 구 목록
		api.GET("/districts", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"districts": []string{"강남구", "마포구"},
			})
		})
	}

	return r
}
