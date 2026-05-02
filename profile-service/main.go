package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type profileResponse struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Level    string `json:"level"`
	Service  string `json:"service"`
}

func main() {
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
		c.JSON(400, gin.H{"error": "profile id is required"})
		return
	}

	response := profileResponse{
		ID:       id,
		FullName: "Test User " + id,
		Level:    "basic",
		Service:  "profile-service",
	}

	c.JSON(200, response)
}
