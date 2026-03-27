package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler godoc
// @Summary Health check endpoint
// @Description Check if the API server is running properly
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func HealthHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		service := Health{}
		res := service.Status()
		c.JSON(http.StatusOK, res)
	}
}

