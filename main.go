package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/receipts/:id/points", getReceipt)
	router.POST("/receipts/process", processReceipt)
	router.Run("localhost:8080")
}
