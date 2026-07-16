package main

import (
	"database/sql"
	"errors"
	"fmt"
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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config := appConfig.ParseFlags()
	if config.DatabaseURL == "" && config.FilePath == "" {
		log.Fatal("configuration error: neither database URL nor file path is provided")
	}

	var repo model.ShortenerRepository
	var dbConn *sql.DB
	var err error
	if config.DatabaseURL != "" {
		dbConn, err = initDatabase(config.DatabaseURL)
		if err != nil {
			log.Fatalf("Error initializing database connection: %v", err)
		}
		repo = dbrepo.NewURLRepository(dbConn)
	} else {
		repo, err = filerepo.NewURLRepository(config.FilePath)
		if err != nil {
			log.Fatalf("failed to init file repository: %v", err)
		}
	}

	mngr, err := initAuditManager(config)
	if err != nil {
		log.Fatalf("failed to init audit: %v", err)
	}

	shortenerService := service.NewShortenerService(repo, config.MaxWorkers)

	h := handler.NewHandler(shortenerService, dbConn, mngr)
	restAPIHandler := handler.NewRestAPIHandler(shortenerService, mngr)

	r := setupRouter(h, restAPIHandler, config)

	logAppInfo()

	if config.EnableHTTPS {
		err = r.RunTLS(config.ServerAddr, config.CertFile, config.KeyFile)
	} else {
		err = r.Run(config.ServerAddr)
	}
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func migrateDB(conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{
		SchemaName: "public",
	})
	if err != nil {
		return err
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "..", "..", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}
	return nil
}

func initDatabase(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	err = migrateDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return db, nil
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
	if config.AuditURL != "" {
		mgr.AddReceiver(audit.NewRemoteReceiver(config.AuditURL))
	}
	return mgr, nil
}
func setupRouter(h handler.Handler, restAPIHandler handler.RestAPIHandler, config appConfig.Config) *gin.Engine {
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

	return r
}

func logAppInfo() {
	fmt.Printf("Build version: %s\n", getOrDefault(buildVersion, "N/A"))
	fmt.Printf("Build date: %s\n", getOrDefault(buildDate, "N/A"))
	fmt.Printf("Build commit: %s\n", getOrDefault(buildCommit, "N/A"))
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
