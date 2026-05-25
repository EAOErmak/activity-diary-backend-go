package main

import (
	"log"
	"net/http"
	"time"

	"activitydiary/api/internal/analyticsclient"
	"activitydiary/api/internal/auth"
	"activitydiary/api/internal/config"
	"activitydiary/api/internal/db"
	"activitydiary/api/internal/diary"
	"activitydiary/api/internal/dictionary"
	"activitydiary/api/internal/foods"
	"activitydiary/api/internal/metriclinks"
	"activitydiary/api/internal/middleware"
	"activitydiary/api/internal/tagcharttypes"
	"activitydiary/api/internal/tags"
	"activitydiary/api/internal/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database, sqlDB, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	if err := db.RunMigrations(cfg); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	if err := db.SeedInitialData(database); err != nil {
		log.Fatalf("seed data: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	authService := auth.NewService(database, jwtManager)
	userService := users.NewService(database)
	tagService := tags.NewService(database)
	dictionaryService := dictionary.NewService(database)
	diaryService := diary.NewService(database, tagService)
	metricLinkService := metriclinks.NewService(database)
	tagChartTypeService := tagcharttypes.NewService(database)
	foodService := foods.NewService(database)
	analyticsHTTP := analyticsclient.New(cfg.AnalyticsServiceURL)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5174", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")

	authHandler := auth.NewHandler(authService)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.Refresh)

	api.GET("/tags", tags.NewHandler(tagService).ListPublic)
	api.GET("/tags/metrics", tags.NewHandler(tagService).ListMetricsByTags)
	api.GET("/dictionary/metric-names/:metricNameId/units", dictionary.NewHandler(dictionaryService).ListMetricUnits)
	api.GET("/dict/all", dictionary.NewHandler(dictionaryService).ListAll)
	api.GET("/general-foods", foods.NewHandler(foodService).ListGeneralFoods)
	api.GET("/general-foods/:id", foods.NewHandler(foodService).GetGeneralFood)

	protected := api.Group("/")
	protected.Use(middleware.Auth(jwtManager))
	{
		userHandler := users.NewHandler(userService)
		protected.GET("/user/me", userHandler.Me)

		diaryHandler := diary.NewHandler(diaryService)
		protected.GET("/diary/mine", diaryHandler.ListMine)
		protected.GET("/diary/:id", diaryHandler.GetByID)
		protected.POST("/diary", diaryHandler.Create)
		protected.PUT("/diary/:id", diaryHandler.Update)
		protected.DELETE("/diary/:id", diaryHandler.Delete)

		analyticsHandler := analyticsclient.NewHandler(analyticsHTTP)
		protected.GET("/analytics/tags/:tagId/chart-types", analyticsHandler.GetChartTypes)
		protected.GET("/analytics/charts", analyticsHandler.GetChart)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.Auth(jwtManager), middleware.RequireAdmin())
	{
		dictHandler := dictionary.NewHandler(dictionaryService)
		admin.GET("/dict/:type", dictHandler.ListAdmin)
		admin.POST("/dict", dictHandler.CreateAdmin)
		admin.PUT("/dict/:id", dictHandler.UpdateAdmin)
		admin.GET("/dict/search", dictHandler.SearchAdmin)

		metricHandler := metriclinks.NewHandler(metricLinkService)
		admin.GET("/metric-links/metric-name/:metricNameId/units", metricHandler.ListUnitsByMetricName)
		admin.POST("/metric-links", metricHandler.Create)
		admin.DELETE("/metric-links", metricHandler.Delete)

		tagHandler := tags.NewHandler(tagService)
		admin.GET("/tags", tagHandler.ListAdmin)
		admin.POST("/tags", tagHandler.CreateAdmin)
		admin.PUT("/tags/:id", tagHandler.UpdateAdmin)
		admin.POST("/tags/:id/approve", tagHandler.Approve)
		admin.POST("/tags/:id/reject", tagHandler.Reject)
		admin.POST("/tags/:id/deprecate", tagHandler.Deprecate)
		admin.DELETE("/tags/:id", tagHandler.DeleteAdmin)
		admin.GET("/tags/:id/metrics", tagHandler.ListTagMetrics)
		admin.PUT("/tags/:id/metrics", tagHandler.ReplaceTagMetrics)

		tagChartHandler := tagcharttypes.NewHandler(tagChartTypeService)
		admin.GET("/tag-chart-types/tag/:tagId", tagChartHandler.ListByTag)
		admin.POST("/tag-chart-types", tagChartHandler.Create)
		admin.DELETE("/tag-chart-types", tagChartHandler.Delete)

		foodHandler := foods.NewHandler(foodService)
		admin.POST("/general-foods", foodHandler.CreateGeneralFood)
		admin.PUT("/general-foods/:id", foodHandler.UpdateGeneralFood)
		admin.DELETE("/general-foods/:id", foodHandler.DeleteGeneralFood)
	}

	server := &http.Server{
		Addr:              ":18080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("api listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
