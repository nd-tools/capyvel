package middlewareContract

import "github.com/gin-gonic/gin"

type Middleware interface {
	Middleware(ctx *gin.Context)

	MiddlewarePermissions(ctx *gin.Context, permissions []string, requireAll bool) gin.HandlerFunc
}
