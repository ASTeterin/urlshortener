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
	repo, err := file.NewUrlRepository(config.FilePath)
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
	var dbConn *sql.DB
	if config.DbConnStr != "" {
		dbConn, err = sql.Open("pgx", config.DbConnStr)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer dbConn.Close()
		repo = dbrepo.NewUrlRepository(dbConn)
	}
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService, dbConn)
	restApiHandler := handler.NewRestApiHandler(shortenerService)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r := gin.Default()
	r.Use(logger.RequestLogger(), compress.RequestEncoder())
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(ctx, c, config.ResultBaseUrl)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(ctx, c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restApiHandler.GetShortURL(ctx, c, config.ResultBaseUrl)
	})
	r.GET("/ping", func(c *gin.Context) {
		h.CheckDbConnection(ctx, c)
	})
	r.POST("/api/shorten/batch", func(c *gin.Context) {
		restApiHandler.ListShortURLs(ctx, c, config.ResultBaseUrl)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
