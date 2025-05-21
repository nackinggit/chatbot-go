package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func userActionCallback(ctx *gin.Context) {
	fmt.Println("userActionCallback")
}
