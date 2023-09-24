package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func Webhook(storage storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		metricsValue := c.Param("valueM")

		if metricsType == "counter" {
			if value, err := strconv.ParseInt(metricsValue, 10, 64); err == nil {
				storage.UpdateCounter(metricsName, value)
			} else {
				return c.String(http.StatusBadRequest, fmt.Sprintf("%s incorrect values(int) of metric", metricsValue))
			}
		} else if metricsType == "gauge" {
			if value, err := strconv.ParseFloat(metricsValue, 64); err == nil {
				storage.UpdateGauge(metricsName, value)
			} else {
				return c.String(http.StatusBadRequest, fmt.Sprintf("%s incorrect values(float) of metric", metricsValue))
			}
		} else {
			return c.String(http.StatusBadRequest, "Invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}

		c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
		return c.String(http.StatusOK, "")
	}

}

func UpdateJSON(s storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&metric)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}

		switch metric.MType {
		case "counter":
			s.UpdateCounter(metric.ID, *metric.Delta)
		case "gauge":
			s.UpdateGauge(metric.ID, *metric.Value)
		default:
			return c.String(http.StatusNotFound, "Metric not fount or invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, metric)
	}
}

func ValueMetrics(storage storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		if len(storage.GetValue(metricsType, metricsName)) == 0 {
			return c.String(http.StatusNotFound, "")
		}
		err := c.String(http.StatusOK, storage.GetValue(metricsType, metricsName))
		if err != nil {
			return err
		}

		return nil
	}
}

func GetValueJson(s storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&metric)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}
		switch metric.MType {
		case "counter":
			value := s.GetCounterValue(metric.ID)
			metric.Delta = &value
		case "gauge":
			value := s.GetGaugeValue(metric.ID)
			metric.Value = &value
		default:
			return c.String(http.StatusNotFound, "Metric not fount or invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}
		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, metric)
	}
}

func AllMetrics(storage storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		m := storage.AllMetrics()
		result := "Gauge metrics:\n"
		for name, value := range *m.Gauge {
			result += fmt.Sprintf("- %s = %f\n", name, value)
		}

		result += "Counter metrics:\n"
		for name, value := range *m.Counter {
			result += fmt.Sprintf("- %s = %d\n", name, value)
		}
		err := ctx.String(http.StatusOK, result)
		if err != nil {
			return err
		}

		return nil
	}
}
