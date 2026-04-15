package main

import (
	"database/sql"
	"github.com/ASTeterin/urlshortener/internal/compress"
	appConfig "github.com/ASTeterin/urlshortener/internal/config"
	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/model"
	dbrepo "github.com/ASTeterin/urlshortener/internal/repository/db"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	config := appConfig.ParseFlags()
	var dbConn *sql.DB
	var repo model.ShortenerRepository
	var err error
	if config.DBConnStr != "" {
		dbConn, err = sql.Open("pgx", config.DBConnStr)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer dbConn.Close()
		repo = dbrepo.NewURLRepository(dbConn)
	} else {
		repo, err = file.NewURLRepository(config.FilePath)
		if err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService, dbConn)
	restAPIHandler := handler.NewRestAPIHandler(shortenerService)

	r := gin.Default()
	r.Use(logger.RequestLogger(), compress.RequestEncoder())
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c, config.ResultBaseURL)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restAPIHandler.GetShortURL(c, config.ResultBaseURL)
	})
	r.GET("/ping", func(c *gin.Context) {
		h.CheckDBConnection(c)
	})
	r.POST("/api/shorten/batch", func(c *gin.Context) {
		restAPIHandler.ListShortURLs(c, config.ResultBaseURL)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
