package handlers

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Sofja96/go-metrics.git/internal/server/config"
	middleware "github.com/Sofja96/go-metrics.git/internal/server/middleware"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/database"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
)

// APIServer - структура настроек API сервера.
type APIServer struct {
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
}

// New - создает, инициализурет и конфигурирует новый экземпляр ApiServer.
func New(ctx context.Context) *APIServer {
	a := &APIServer{}
	c, err := config.LoadConfig()
	if err != nil {
		log.Printf("error load config: %v", err)
	}

	a.address = c.Address
	a.echo = echo.New()

	var store storage.Storage

	if len(c.DatabaseDSN) == 0 {
		store, err = memory.NewInMemStorage(ctx, c.StoreInterval, c.FilePath, c.Restore)
		if err != nil {
			log.Print(err)
		}
	} else {
		store, err = database.NewStorage(ctx, c.DatabaseDSN)
		if err != nil {
			log.Print(err)
		}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	a.logger = *logger.Sugar()
	a.echo.Use(middleware.WithLogging(a.logger))

	pkFile := c.CryptoKey
	if len(pkFile) != 0 {
		privateKey, err := middleware.LoadPrivateKey(pkFile)
		if err != nil {
			log.Fatalf("Failed to load private key: %v", err)
		}
		a.echo.Use(middleware.DecryptMiddleware(privateKey))
	}

	key := c.HashKey
	if len(key) != 0 {
		a.echo.Use(middleware.HashMacMiddleware([]byte(key)))
	}

	a.echo.Use(middleware.GzipMiddleware())
	a.echo.POST("/update/", UpdateJSON(store))
	a.echo.POST("/updates/", UpdatesBatch(store))
	a.echo.POST("/value/", ValueJSON(store))
	a.echo.GET("/", GetAllMetrics(store))
	a.echo.GET("/value/:typeM/:nameM", ValueMetric(store))
	a.echo.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
	a.echo.GET("/ping", Ping(store))
	return a
}

// Start - запускает сервер на заданном адресе.
func (a *APIServer) Start(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Println("Starting server on", a.address)
		if err := a.echo.Start(a.address); err != nil {
			serverErrors <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, shutting down gracefully...")
	case err := <-serverErrors:
		log.Printf("Error starting server: %v\n", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
		return err
	}

	log.Println("Server shut down gracefully")
	return nil
}
