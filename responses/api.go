package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Api struct {
	TotalRows     int64  `json:"count,omitempty"`
	Relationships any    `json:"relationships,omitempty"`
	QueryParams   any    `json:"queryParams,omitempty"`
	Meta          any    `json:"meta,omitempty"`
	Links         any    `json:"links,omitempty"`
	Message       string `json:"message"`
	Status        int    `json:"status"`
	Data          any    `json:"data"`
	Success       bool   `json:"success"`
}

func (api *Api) Error(ctx *gin.Context, e Error) {
	e.ErrorDetail.LoadDetail()
	e.Status = 400
	e.Success = false
	ctx.JSON(e.Code, e)
	ctx.Abort()
}

func (api *Api) OK(ctx *gin.Context, a Api) {
	var status = 200
	if a.Data == nil {
		status = 204
	}
	a.Status = status
	a.Success = true
	ctx.JSON(http.StatusOK, a)
}
