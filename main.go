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
	router.GET("/api/v1/ad", db.SelectActiveAdvertisements)
	router.Run("localhost:8080")
}
