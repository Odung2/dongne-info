// Package middleware 는 API 서버의 HTTP 미들웨어를 정의한다.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 는 Cross-Origin Resource Sharing 헤더를 설정하는 미들웨어.
//
// Next.js 프론트엔드(localhost:3000, Vercel 도메인)에서
// Go API(localhost:8080, Railway 도메인)를 직접 호출할 수 있도록 허용한다.
//
// MVP에서는 모든 Origin을 허용한다.
// 프로덕션에서는 allowedOrigins를 특정 도메인으로 제한해야 한다.
//
// 사용 예:
//
//	r := gin.Default()
//	r.Use(middleware.CORS())
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
