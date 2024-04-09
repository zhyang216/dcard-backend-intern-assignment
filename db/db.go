package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

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

	// Delete the table first if any
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS advertisements")
	if err != nil {
		log.Printf("Error %s when dropping advertisements table", err)
		return
	}

	// Create the advertisements table
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS advertisements (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			start_at DATETIME NOT NULL,
			end_at DATETIME NOT NULL,
			age_start INT NOT NULL DEFAULT 1,
			age_end INT NOT NULL DEFAULT 100,
			gender VARCHAR(15),
			country VARCHAR(1500),
			platform VARCHAR(255)
		)`)
	if err != nil {
		log.Printf("Error %s when creating advertisements table", err)
		return
	}
	log.Printf("Created tables successfully\n")
}

type Advertisement struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	StartAt  time.Time `json:"startAt"`
	EndAt    time.Time `json:"endAt"`
	AgeStart int       `json:"ageStart"`
	AgeEnd   int       `json:"ageEnd"`
	Gender   []string  `json:"gender"`
	Country  []string  `json:"country"`
	Platform []string  `json:"platform"`
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

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err = db.ExecContext(ctx,
		"INSERT INTO advertisements (title, start_at, end_at, age_start, age_end, gender, country, platform) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		ad.Title, ad.StartAt, ad.EndAt, ad.AgeStart, ad.AgeEnd, strings.Join(ad.Gender, ","), strings.Join(ad.Country, ","), strings.Join(ad.Platform, ","))
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return
	}
	log.Printf("Inserted advertisement %s successfully\n", ad.Title)
	c.JSON(200, gin.H{"message": "Inserted advertisement successfully"})
}

func SelectActiveAdvertisements(c *gin.Context) {
	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return
	}
	defer db.Close()

	offset := c.DefaultQuery("offset", "0")
	limit := c.DefaultQuery("limit", "5")
	age_lb := c.DefaultQuery("age", "100")
	age_ub := c.DefaultQuery("age", "1")
	gender := c.DefaultQuery("gender", "")
	country := c.DefaultQuery("country", "")
	platform := c.DefaultQuery("platform", "")

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	rows, err := db.QueryContext(ctx,
		`SELECT * FROM advertisements
		WHERE (age_start <= ? AND age_end >= ?)
		AND (gender LIKE ? OR gender IS NULL)
		AND (country LIKE ? OR country IS NULL)
		AND (platform LIKE ? OR platform IS NULL)
		ORDER BY end_at ASC
		LIMIT ? OFFSET ?`,
		age_lb, age_ub, gender, country, platform, limit, offset)
	if err != nil {
		log.Printf("Error %s when querying active advertisements", err)
		return
	}
	defer rows.Close()

	ads := make([]Advertisement, 0)
	for rows.Next() {
		var ad Advertisement
		if err := rows.Scan(&ad.Title, &ad.EndAt); err != nil {
			log.Printf("Error %s when scanning rows", err)
			return
		}
		ads = append(ads, ad)
	}
	c.JSON(200, ads)
}
