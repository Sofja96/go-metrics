package handlers

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func Webhook(metrics storage.Metrics) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		metricsValue := c.Param("valueM")

		if metricsType == "counter" {
			if value, err := strconv.ParseInt(metricsValue, 10, 64); err == nil {
				metrics.UpdateCounter(metricsName, value)
			} else {
				return c.String(http.StatusBadRequest, fmt.Sprintf("%s incorrect values(int) of metric", metricsValue))
			}
		} else if metricsType == "gauge" {
			if value, err := strconv.ParseFloat(metricsValue, 64); err == nil {
				metrics.UpdateGauge(metricsName, value)
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

func ValueMetrics(metrics storage.Metrics) echo.HandlerFunc {
	return func(c echo.Context) error {
		metricsType := c.Param("typeM")
		metricsName := c.Param("nameM")
		if len(metrics.GetValue(metricsType, metricsName)) == 0 {
			return c.String(http.StatusNotFound, "")
		}
		err := c.String(http.StatusOK, metrics.GetValue(metricsType, metricsName))
		if err != nil {
			return err
		}

		return nil
	}
}

func AllMetrics(metrics storage.Metrics) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		m := metrics.AllMetrics()
		result := "Gauge metrics:\n"
		for name, value := range m.Gauge {
			result += fmt.Sprintf("- %s = %f\n", name, value)
		}

		result += "Counter metrics:\n"
		for name, value := range m.Counter {
			result += fmt.Sprintf("- %s = %d\n", name, value)
		}
		err := ctx.String(http.StatusOK, result)
		if err != nil {
			return err
		}

		return nil
	}
}
