package utils

import "github.com/gin-gonic/gin"

func SetUpRouter(setHandlersFnc func(r *gin.Engine)) *gin.Engine {
	router := gin.Default()
	setHandlersFnc(router)
	return router
}
