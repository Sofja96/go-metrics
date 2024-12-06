package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

// Webhook - обработчик для обновления одной метрики.
func Webhook(storage storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		metricsValue := c.Param("valueM")

		if metricsType == "counter" {
			if value, err := strconv.ParseInt(metricsValue, 10, 64); err == nil {
				_, err := storage.UpdateCounter(metricsName, value)
				if err != nil {
					return err
				}
			} else {
				return c.String(http.StatusBadRequest, "incorrect values(int) of metric: "+metricsValue)
			}
		} else if metricsType == "gauge" {
			if value, err := strconv.ParseFloat(metricsValue, 64); err == nil {
				_, err := storage.UpdateGauge(metricsName, value)
				if err != nil {
					return err
				}
			} else {
				return c.String(http.StatusBadRequest, "incorrect values(float) of metric: "+metricsValue)
			}
		} else {
			return c.String(http.StatusBadRequest, "Invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.String(http.StatusOK, "")
	}

}

// UpdateJSON - обработчик для обновления одной метрики в формате JSON.
func UpdateJSON(s storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Error in JSON decode: "+err.Error())
		}
		if len(metric.ID) == 0 {
			return ctx.String(http.StatusNotFound, "No id metric for"+metric.MType)
		}
		switch metric.MType {
		case "counter":
			_, err := s.UpdateCounter(metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		case "gauge":
			_, err := s.UpdateGauge(metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		default:
			return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
		}
		ctx.Response().Header().Set("Content-Type", "application/json")
		if ctx.Request().Header.Get("Content-Type") != "application/json" {
			return ctx.String(http.StatusUnsupportedMediaType, "")
		}
		return ctx.JSON(http.StatusOK, metric)
	}
}

// UpdatesBatch - обработчик для обновления нескольких метрик.
func UpdatesBatch(s storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metrics []models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metrics)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Error in JSON decode"+err.Error())
		}
		if len(metrics) == 0 {
			return ctx.String(http.StatusBadRequest, "metric is empty")
		}
		err = s.BatchUpdate(ctx.Response().Writer, metrics)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "")
		}
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(ctx.Response().Writer).Encode(metrics)
	}
}

// ValueMetric - обработчик для получения метрики по типу и имени.
func ValueMetric(storage storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		var v string
		switch metricsType {
		case "counter":
			value, ok := storage.GetCounterValue(metricsName)
			if !ok {
				return c.String(http.StatusNotFound, "")
			}
			v = fmt.Sprint(value)
		case "gauge":
			value, ok := storage.GetGaugeValue(metricsName)
			if !ok {
				return c.String(http.StatusNotFound, "")
			}
			v = fmt.Sprint(value)
		default:
			return c.String(http.StatusNotFound, "Metric not fount or invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}
		c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
		return c.String(http.StatusOK, v)
	}
}

// ValueJSON - обработчик для получения метрики по типу и имени в формате JSON.
func ValueJSON(s storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&metric)
		if err != nil {
			return c.String(http.StatusBadRequest, "Error in JSON decode: "+err.Error())
		}
		if len(metric.ID) == 0 {
			return c.String(http.StatusNotFound, "No id metric for: "+metric.MType)
		}
		switch metric.MType {
		case "counter":
			value, ok := s.GetCounterValue(metric.ID)
			if !ok {
				return c.String(http.StatusNotFound, "")
			}
			metric.Delta = &value
		case "gauge":
			value, ok := s.GetGaugeValue(metric.ID)
			if !ok {
				return c.String(http.StatusNotFound, "")
			}
			metric.Value = &value
		default:
			return c.String(http.StatusBadRequest, "Metric not found or invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}
		c.Response().Header().Set("Content-Type", "application/json")
		if c.Request().Header.Get("Content-Type") != "application/json" {
			return c.String(http.StatusUnsupportedMediaType, "")
		}
		return c.JSON(http.StatusOK, metric)
	}
}

// GetAllMetrics - обработчик для получения всех метрик.
func GetAllMetrics(storage storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/html")
		gaugeMetrics, err := storage.GetAllGauges()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "")
		}
		counterMetrics, err := storage.GetAllCounters()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "")
		}

		var result strings.Builder

		// Формируем строку с метриками типа gauge
		result.WriteString("<html><body>")
		result.WriteString("<h2>Gauge metrics:</h2>")
		result.WriteString("<ul>")
		for _, metric := range gaugeMetrics {
			result.WriteString(fmt.Sprintf("<li>%s = %.2f</li>", metric.Name, metric.Value))
		}
		result.WriteString("</ul>")

		// Формируем строку с метриками типа counter
		result.WriteString("<h2>Counter metrics:</h2>")
		result.WriteString("<ul>")
		for _, metric := range counterMetrics {
			result.WriteString(fmt.Sprintf("<li>%s = %d</li>", metric.Name, metric.Value))
		}
		result.WriteString("</ul>")
		result.WriteString("</body></html>")

		return ctx.String(http.StatusOK, result.String())

	}
}

// Ping - обработчик для определения доступности БД.
func Ping(storage storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/html")
		err := storage.Ping()
		if err == nil {
			ctx.String(http.StatusOK, "Connection database is OK")
		} else {
			ctx.String(http.StatusInternalServerError, "Connection database is NOT ok")
		}
		if err != nil {
			return fmt.Errorf("error connection db: %w", err)
		}
		return nil
	}
}
