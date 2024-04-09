package main

import (
	db "dcard/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.DbConnect()
	db.CreateTable()
	router := gin.Default()
	router.POST("/api/v1/ad", db.InsertAdvertisement)
	// router.GET("/api/v1/ad", select_active_advertisements)
	router.Run("localhost:8080")
}
