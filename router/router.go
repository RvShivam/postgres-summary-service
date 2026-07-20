package router

import (
	"github.com/RvShivam/postgres-summary-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func New(handler *handler.Handler) *gin.Engine {

	router := gin.Default()

	router.POST("/summary/sync", handler.SyncSummary)
	router.GET("/summaries", handler.ListSummaries)
	router.GET("/summaries/:id", handler.GetSummary)

	return router
}
