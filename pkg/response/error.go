package response

import (
	"github.com/gin-gonic/gin"

	"shortbin/pkg/config"
)

const (
	InvalidParameters  = "invalid parameters"
	SomethingWentWrong = "something went wrong"
	Unauthorized       = "unauthorized"
	IdNotFound         = "id not found"
	UserAlreadyExists  = "user already exists"
	EmptyUserId
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
