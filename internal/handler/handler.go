package handler

import (
	"github.com/RvShivam/postgres-summary-service/internal/service"
)

type Handler struct {
	service service.Service
}

func New(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}
