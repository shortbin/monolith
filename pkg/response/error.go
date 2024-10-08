package response

import (
	"github.com/gin-gonic/gin"

	"shortbin/pkg/config"
)

const (
	InvalidParameters  = "invalid parameters"
	SomethingWentWrong = "something went wrong"
	//nolint:gosec
	WrongCredentials   = "wrong credentials"
	Unauthorized       = "unauthorized"
	IDNotFound         = "id not found"
	UserAlreadyExists  = "user already exists"
	EmptyUserID        = "user id is empty"
	IDLengthNotInRange = "id length not in range"
	UserNotFound       = "user not found"
	NoRowsInResultSet  = "no rows in result set"
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
