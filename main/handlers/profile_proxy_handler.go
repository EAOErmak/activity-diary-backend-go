package handlers

import (
	"net/http"

	"go-learn/main/clients"

	"github.com/gin-gonic/gin"
)

var profileClient = clients.NewProfileClient("http://localhost:8081")

func GetExternalProfile(c *gin.Context) {
	profileID := c.Param("id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile id is required"})
		return
	}

	profile, err := profileClient.GetProfileByID(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
