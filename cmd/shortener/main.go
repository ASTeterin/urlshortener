package main

import (
	"log"

	appConfig "github.com/ASTeterin/urlshortener/internal/config"
	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	repo := repository.NewUrlRepository()
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService)
	restApiHandler := handler.NewRestApiHandler(shortenerService)
	config := appConfig.ParseFlags()

	r := gin.Default()
	r.Use(logger.RequestLogger())
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c, config.ResultBaseUrl)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restApiHandler.GetShortURL(c, config.ResultBaseUrl)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
