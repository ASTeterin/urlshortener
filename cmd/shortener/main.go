package main

import (
	"log"

	"github.com/ASTeterin/urlshortener/internal/compress"
	appConfig "github.com/ASTeterin/urlshortener/internal/config"
	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	config := appConfig.ParseFlags()
	repo, err := repository.NewUrlRepository(config.FilePath)
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService, config.DbConnStr)
	restApiHandler := handler.NewRestApiHandler(shortenerService)

	r := gin.Default()
	r.Use(logger.RequestLogger(), compress.RequestEncoder())
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c, config.ResultBaseUrl)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restApiHandler.GetShortURL(c, config.ResultBaseUrl)
	})
	r.GET("/ping", func(c *gin.Context) {
		h.CheckDbConnection(c)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
