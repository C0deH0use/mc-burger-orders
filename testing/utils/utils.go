package utils

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func SetUpRouter(setHandlersFnc func(r *gin.Engine)) *gin.Engine {
	router := gin.Default()
	setHandlersFnc(router)
	return router
}

func IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
