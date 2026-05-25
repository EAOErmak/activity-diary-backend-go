package main

import (
	"log"
	"net/http"
	"time"

	"activitydiary/analytics/internal/config"
	"activitydiary/analytics/internal/db"
	"activitydiary/analytics/internal/handlers"
	"activitydiary/analytics/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database, sqlDB, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open analytics db: %v", err)
	}
	defer sqlDB.Close()

	svc := service.New(database)
	handler := handlers.New(svc)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/internal/tags/:tagId/chart-types", handler.GetChartTypes)
	router.GET("/internal/charts", handler.GetChart)

	server := &http.Server{
		Addr:              ":18081",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("analytics listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
