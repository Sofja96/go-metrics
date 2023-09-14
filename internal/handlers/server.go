package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
)

func CreateServer(s *storage.MemStorage) *echo.Echo {
	e := echo.New()
	e.GET("/", AllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetrics(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	return e
}
