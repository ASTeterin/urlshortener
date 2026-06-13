package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/ASTeterin/urlshortener/internal/audit"
	"github.com/ASTeterin/urlshortener/internal/compress"
	appConfig "github.com/ASTeterin/urlshortener/internal/config"
	"github.com/ASTeterin/urlshortener/internal/cookie"
	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/model"
	dbrepo "github.com/ASTeterin/urlshortener/internal/repository/db"
	filerepo "github.com/ASTeterin/urlshortener/internal/repository/file"
	"github.com/ASTeterin/urlshortener/internal/service"
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
		migrateDB(dbConn)
		repo = dbrepo.NewURLRepository(dbConn)
	} else {
		repo, err = filerepo.NewURLRepository(config.FilePath)
		if err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}

	mngr, err := initAuditManager(config)
	if err != nil {
		log.Fatalf("failed to init audit: %v", err)
	}

	shortenerService := service.NewShortenerService(repo, config.MaxWorkers)
	h := handler.NewHandler(shortenerService, dbConn, mngr)
	restAPIHandler := handler.NewRestAPIHandler(shortenerService)

	r := gin.Default()
	r.Use(logger.RequestLogger(), compress.RequestEncoder(), cookie.CookieHandler(config.SigningKey))
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
	r.GET("/api/user/urls", func(c *gin.Context) {
		restAPIHandler.ListUserURLs(c, config.ResultBaseURL)
	})
	r.DELETE("/api/user/urls", func(c *gin.Context) {
		restAPIHandler.BatchRemove(c)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func migrateDB(conn *sql.DB) {
	driver, err := postgres.WithInstance(conn, &postgres.Config{
		SchemaName: "public",
	})
	if err != nil {
		log.Fatal(err)
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "..", "..", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Migrations directory not found: %s", migrationsPath)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	}
}

func initAuditManager(config appConfig.Config) (*audit.Manager, error) {
	mgr := audit.NewAuditManager()

	if config.AuditFilePath != "" {
		r, err := audit.NewFileReceiver(config.AuditFilePath)
		if err != nil {
			return nil, err
		}
		mgr.AddReceiver(r)
	}
	if config.AuditUrl != "" {
		mgr.AddReceiver(audit.NewRemoteReceiver(config.AuditUrl))
	}
	return mgr, nil
}
