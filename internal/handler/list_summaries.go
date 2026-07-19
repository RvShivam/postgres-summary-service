package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListSummaries(c *gin.Context) {

	summaries, err := h.service.ListSummaries(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summaries)
}
