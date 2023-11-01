package middleware

import "github.com/gin-gonic/gin"

type EndpointsSetup interface {
	Setup(r *gin.Engine)
}
