package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	config "welcome-to-the-jungle/pkg/config"
	db "welcome-to-the-jungle/pkg/db"
	_ "welcome-to-the-jungle/pkg/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	DBClient *db.DBClient
	appCache *cache.Cache
)

// @title Welcome to the Jungle - Client API
// @version 1.0
// @description This API provides endpoints to manage client data including loading data from a CSV file into PostgreSQL, and retrieving client information by UID, filtering by age range, and searching by name. The service is built with Golang using the Gin framework and provides Swagger documentation for easy API exploration.
// @host localhost:8080
// @BasePath /
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading configuration")
	}

	// Initialize cache with a default expiration time of 5 minutes, and purging every 10 minutes
	appCache = cache.New(5*time.Minute, 10*time.Minute)

	// Connect to the database with retries
	DBClient, err = db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	// Initialize the database schema
	err = DBClient.InitSchema()
	if err != nil {
		log.Fatal("Error initializing database schema:", err)
	}

	// Check if data is already loaded
	if !cfg.DataLoaded {
		// Load CSV data into the database
		err := DBClient.LoadCSV(cfg.CSVFilePath)
		if err != nil {
			log.Fatal("Error loading CSV data into the database:", err)
		}

		// Optimize the database
		err = DBClient.OptimizeDatabase()
		if err != nil {
			log.Fatal("Error optimizing the database:", err)
		}
	}

	// Using release mode to be faster
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/info", getClientInfo)
	r.GET("/info/by-age", getClientsByAge)
	r.GET("/search", searchClientsByName)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}

// @Summary Get client info
// @Description Get client info by UID
// @Produce json
// @Param uid query int true "Client UID"
// @Success 200 {object} models.Client
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /info [get]
func getClientInfo(c *gin.Context) {
	uidStr := c.Query("uid")
	if uidStr == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "UID is required"})
		return
	}

	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UID"})
		return
	}

	// Check if the data is in cache
	if cachedClient, found := appCache.Get(uidStr); found {
		c.JSON(http.StatusOK, cachedClient)
		return
	}

	client, err := DBClient.GetClientByUID(uid)
	if err != nil {
		log.Printf(err.Error())
		c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
		return
	}

	// Cache the client data
	appCache.Set(uidStr, client, cache.DefaultExpiration)

	c.JSON(http.StatusOK, client)
}

// @Summary Get clients by age
// @Description Get clients by age range
// @Produce json
// @Param start query string true "Start Date" format(date) example(1970-01-01)
// @Param end query string true "End Date" format(date) example(1980-01-01)
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Client
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /info/by-age [get]
func getClientsByAge(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Start and end dates are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid start date"})
		return
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid end date"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100 // default limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // default offset
	}

	clients, err := DBClient.GetClientByAgeWithPagination(start, end, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// @Summary Search clients by name
// @Description Search clients by name
// @Produce json
// @Param name query string true "Name"
// @Success 200 {array} models.Client
// @Failure 500 {object} map[string]string
// @Router /search [get]
func searchClientsByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
		return
	}

	cacheKey := "name-" + name
	// Check if the data is in cache
	if cachedClients, found := appCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cachedClients)
		return
	}

	clients, err := DBClient.SearchClientByName(name)
	if err != nil {
		log.Printf(err.Error())
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
		return
	}

	// Cache the clients data
	appCache.Set(cacheKey, clients, cache.DefaultExpiration)

	c.JSON(http.StatusOK, clients)
}
