package middlewareContract

import "github.com/gin-gonic/gin"

type Middleware interface {
	Middleware(ctx *gin.Context)
}

type MiddlewarePermissions interface {
	MiddlewarePermissions(ctx *gin.Context, permissions []string, requireAll bool) gin.HandlerFunc
}
