package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/biter777/countries"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const (
	username = "root"
	password = "admin"
	hostname = "127.0.0.1:3306"
	dbname   = "dcard"
)

func dsn(dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
}

func DbConnect() {
	db, err := sql.Open("mysql", dsn(""))
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+dbname)
	if err != nil {
		log.Printf("Error %s when creating DB\n", err)
		return
	}
	no, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows", err)
		return
	}
	log.Printf("rows affected: %d\n", no)
	db.Close()

	db, err = sql.Open("mysql", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return
	}
	defer db.Close()

	// Set up the connection pool
	// Set the maximum number of open (in-use + idle) connections in the pool
	db.SetMaxOpenConns(20)

	// Set the maximum number of idle connections in the pool
	db.SetMaxIdleConns(20)

	// Set the maximum lifetime of a connection in the pool
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Errors %s pinging DB", err)
		return
	}
	log.Printf("Connected to DB %s successfully\n", dbname)
}

func CreateTable() {
	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	// Extract all countries from the countries package
	countriesString := "["
	for _, country := range countries.All() {
		countriesString += fmt.Sprintf("\"%s\",", country.Alpha2())
	}
	countriesString = strings.TrimSuffix(countriesString, ",")
	countriesString += "]"

	// Create the advertisements table
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS advertisements (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			start_at DATETIME NOT NULL,
			end_at DATETIME NOT NULL,
			age_from INT NOT NULL DEFAULT 1,
			age_to INT NOT NULL DEFAULT 100,
			gender VARCHAR(15) NOT NULL DEFAULT '["M", "F"]',
			countries VARCHAR(1500) NOT NULL DEFAULT + `+countriesString+`,
			platforms VARCHAR(255) NOT NULL DEFAULT '["iOS", "Android", "Web"]',
		)`)
	if err != nil {
		log.Printf("Error %s when creating advertisements table", err)
		return
	}
	log.Printf("Created tables successfully\n")
}

type Advertisement struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	StartAt   string `json:"startAt"`
	EndAt     string `json:"endAt"`
	Condition struct {
		AgeStart int    `json:"ageStart"`
		AgeEnd   int    `json:"ageEnd"`
		Gender   string `json:"gender"`
		Country  string `json:"country"`
		Platform string `json:"platform"`
	} `json:"condition"`
}

func InsertAdvertisement(c *gin.Context) {
	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return
	}
	defer db.Close()

	var ad Advertisement
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	print(ad.Title)
	print(ad.StartAt)
	print(ad.EndAt)
	print(ad.Condition.AgeStart)
	print(ad.Condition.AgeEnd)
	print(ad.Condition.Gender)
	print(ad.Condition.Country)
	print(ad.Condition.Platform)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		`INSERT INTO advertisements (title, start_at, end_at, age_from, age_to, gender, countries, platforms) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ad.Title, ad.StartAt, ad.EndAt, ad.Condition.AgeStart, ad.Condition.AgeEnd, ad.Condition.Gender, ad.Condition.Country, ad.Condition.Platform)
	if err != nil {
		log.Printf("Failed to insert advertisement: %s", err)
		return
	}
	log.Printf("Inserted advertisement successfully\n")
	c.JSON(200, gin.H{"message": "Inserted advertisement successfully"})
}

// func select_active_advertisements(c *gin.Context) {
// 	db, err := sql.Open("mysql", dsn(dbname))
// 	if err != nil {
// 		log.Printf("Error %s when opening DB", err)
// 		return
// 	}
// 	defer db.Close()

// 	offset := c.Query("offset")
// 	limit := c.DefaultQuery("limit", "5")
// 	age := c.Query("age")
// 	gender := c.Query("gender")
// 	country := c.Query("country")
// 	platform := c.Query("platform")

// 	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancelfunc()

// 	query := `SELECT a.* FROM advertisements a
// 	LEFT JOIN advertisement_gender ag ON a.id = ag.advertisement_id
// 	LEFT JOIN advertisement_country ac ON a.id = ac.advertisement_id
// 	LEFT JOIN advertisement_platform ap ON a.id = ap.advertisement_id
// 	WHERE a.start_at <= NOW() AND a.end_at >= NOW()
// 	AND a.age_from <= ? AND a.age_to >= ?
// 	AND ag
// }
