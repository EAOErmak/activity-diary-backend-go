package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go-learn/main/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type profileResponse struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Level    string `json:"level"`
	Service  string `json:"service"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatalf("profile-service database is unavailable: %v", err)
	}

	router := gin.Default()
	router.GET("/profile/:id", handleProfile)

	log.Println("profile-service listening on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("profile-service failed: %v", err)
	}
}

func handleProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile id is required"})
		return
	}

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile id must be a number"})
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	response := profileResponse{
		ID:       strconv.FormatUint(uint64(user.ID), 10),
		FullName: buildFullName(user.Username),
		Level:    profileLevel(user.Role),
		Service:  "profile-service",
	}

	c.JSON(http.StatusOK, response)
}

func connectDB() (*gorm.DB, error) {
	dsn := databaseDSN()
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func databaseDSN() string {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5432")
	user := envOrDefault("DB_USER", "postgres")
	password := envOrDefault("DB_PASSWORD", "postgres")
	name := envOrDefault("DB_NAME", "postgres")
	sslMode := envOrDefault("DB_SSLMODE", "disable")

	return "host=" + host +
		" user=" + user +
		" password=" + password +
		" dbname=" + name +
		" port=" + port +
		" sslmode=" + sslMode
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func buildFullName(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "Unknown User"
	}

	parts := strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(username))
	for i := range parts {
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}

	return strings.Join(parts, " ")
}

func profileLevel(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "advanced"
	default:
		return "basic"
	}
}
