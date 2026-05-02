package main

import (
	"go-learn/main/database"
	"go-learn/main/handlers"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Printf("diary database is unavailable: %v", err)
	}

	router := gin.Default()

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)
	router.GET("/external-profile/:id", handlers.GetExternalProfile)

	dictionary := router.Group("/dictionary-items")
	dictionary.Use(handlers.AuthMiddleware())
	{
		dictionary.GET("", handlers.GetAllDictionaryItems)
		dictionary.POST("", handlers.CreateDictionaryItem)
		dictionary.GET("/:id", handlers.GetDictionaryItemByID)
		dictionary.PUT("/:id", handlers.UpdateDictionaryItem)
		dictionary.DELETE("/:id", handlers.DeleteDictionaryItem)
	}

	diary := router.Group("/diary")
	diary.Use(handlers.AuthMiddleware())
	{
		diary.GET("", handlers.GetAllMineDiaryEntries)
		diary.GET("/all", handlers.GetAllDiaryEntriesForAllUsers)
		diary.POST("", handlers.CreateDiaryEntry)
		diary.GET("/:id", handlers.GetDiaryEntryByID)
		diary.PUT("/:id", handlers.UpdateDiaryEntry)
		diary.DELETE("/:id", handlers.DeleteDiaryEntry)
	}

	port := strings.TrimSpace(os.Getenv("APP_PORT"))
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("main service failed: %v", err)
	}
}
