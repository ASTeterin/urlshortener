package main

import (
	"context"
	"database/sql"
	dbrepo "github.com/ASTeterin/urlshortener/internal/repository/db"
	"log"
	"time"

	"github.com/ASTeterin/urlshortener/internal/compress"
	appConfig "github.com/ASTeterin/urlshortener/internal/config"
	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	config := appConfig.ParseFlags()
	repo, err := file.NewURLRepository(config.FilePath)
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
	var dbConn *sql.DB
	if config.DBConnStr != "" {
		dbConn, err = sql.Open("pgx", config.DBConnStr)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer dbConn.Close()
		repo = dbrepo.NewURLRepository(dbConn)
	}
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService, dbConn)
	restAPIHandler := handler.NewRestAPIHandler(shortenerService)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r := gin.Default()
	r.Use(logger.RequestLogger(), compress.RequestEncoder())
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(ctx, c, config.ResultBaseURL)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(ctx, c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restAPIHandler.GetShortURL(ctx, c, config.ResultBaseURL)
	})
	r.GET("/ping", func(c *gin.Context) {
		h.CheckDBConnection(ctx, c)
	})
	r.POST("/api/shorten/batch", func(c *gin.Context) {
		restAPIHandler.ListShortURLs(ctx, c, config.ResultBaseURL)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
