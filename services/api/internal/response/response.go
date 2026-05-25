package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Success bool        `json:"success"`
	Message *string     `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Envelope{
		Success: true,
		Message: nil,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Message: &message,
		Data:    nil,
	})
}
