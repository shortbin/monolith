package response

import (
	"github.com/gin-gonic/gin"

	"shortbin/pkg/config"
)

func Error(c *gin.Context, status int, err error, message string) {
	errorRes := map[string]interface{}{
		"message": message,
	}

	if config.GetConfig().Environment != config.ProductionEnv {
		errorRes["debug"] = err.Error()
	}

	c.JSON(status, errorRes)
	// c.JSON(status, Response{Error: errorRes})
}
