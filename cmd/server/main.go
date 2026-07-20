package main

import (
	"log"

	"github.com/RvShivam/postgres-summary-service/internal/config"
	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/RvShivam/postgres-summary-service/internal/handler"
	"github.com/RvShivam/postgres-summary-service/internal/local"
	"github.com/RvShivam/postgres-summary-service/internal/repository"
	"github.com/RvShivam/postgres-summary-service/internal/service"
	"github.com/RvShivam/postgres-summary-service/router"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := local.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)

	client := external.NewClient(cfg.External.BaseURL)

	svc := service.New(repo, client)

	h := handler.New(svc)

	r := router.New(h)

	log.Printf("server listening on :%s", cfg.Server.Port)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
