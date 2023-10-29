package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/middleware"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func CreateServer(cfg *config.Config, s storage.Storage) *echo.Echo {
	//	c := config.LoadConfig()
	var sugar zap.SugaredLogger
	//var key string
	//key := config.LoadConfig().HashKey
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()
	e := echo.New()
	e.Use(middleware.WithLogging(sugar))
	e.Use(middleware.GzipMiddleware())
	e.POST("/update/", UpdateJSON(s))
	e.POST("/updates/", UpdatesBatch(s))
	e.POST("/value/", ValueJSON(s))
	e.GET("/", GetAllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetric(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	e.GET("/ping", Ping(s))
	//e.Use(middleware.HashMacMiddleware(([]byte(key))))
	e.Use(middleware.HashMacMiddleware(cfg.HashKey))
	//if len(c.HashKey) > 0 {
	//	e.Use(middleware.HashMacMiddleware([]byte(c.HashKey)))
	//}
	return e
}

//func CreateServerHash(s storage.Storage, key string) *echo.Echo {
//	var sugar zap.SugaredLogger
//	//var key string
//	logger, err := zap.NewDevelopment()
//	if err != nil {
//		panic(err)
//	}
//	defer logger.Sync()
//	sugar = *logger.Sugar()
//	e := echo.New()
//	e.Use(middleware.WithLogging(sugar))
//	e.Use(middleware.GzipMiddleware())
//	if len(key) > 0 {
//		e.Use(middleware.HashMacMiddleware([]byte(key)))
//	}
//	e.POST("/update/", UpdateJSON(s))
//	e.POST("/updates/", UpdatesBatch(s))
//	e.POST("/value/", ValueJSON(s))
//	e.GET("/", GetAllMetrics(s))
//	e.GET("/value/:typeM/:nameM", ValueMetric(s))
//	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
//	e.GET("/ping", Ping(s))
//	return e
//}
