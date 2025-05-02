package middlewareContract

import "github.com/gin-gonic/gin"

type Middleware interface {
	Middleware(ctx *gin.Context)
}

type MiddlewarePermissions interface {
	MiddlewarePermissions(permissions []string, requireAll bool) gin.HandlerFunc
}
