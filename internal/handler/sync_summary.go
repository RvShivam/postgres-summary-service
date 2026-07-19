package handler

import (
	"net/http"

	"github.com/RvShivam/postgres-summary-service/internal/service"
	"github.com/gin-gonic/gin"
)

type SyncSummaryRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	DBName   string `json:"dbname" binding:"required"`
}

func (h *Handler) SyncSummary(c *gin.Context) {

	var req SyncSummaryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	summary, err := h.service.SyncSummary(
		c.Request.Context(),
		service.SyncRequest{
			Host:     req.Host,
			Port:     req.Port,
			User:     req.User,
			Password: req.Password,
			DBName:   req.DBName,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, summary)
}
