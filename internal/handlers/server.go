package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/middleware"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func CreateServer(s *storage.MemStorage) *echo.Echo {
	var sugar zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = *logger.Sugar()
	e := echo.New()
	e.Use(middleware.WithLogging(sugar))
	e.GET("/", AllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetrics(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	return e
}
