package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"log"
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

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.String(http.StatusOK, "")
	}

}

func UpdateJSON(s storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}

		switch metric.MType {
		case "counter":
			s.UpdateCounter(metric.ID, *metric.Delta)
		case "gauge":
			s.UpdateGauge(metric.ID, *metric.Value)
		default:
			return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
		}
		//encoder := json.NewEncoder(ctx.Response().Writer)
		//err = encoder.Encode(&metric)
		//if err != nil {
		//	return ctx.String(http.StatusInternalServerError, "")
		//}

		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.JSON(http.StatusOK, metric)
	}
}

//func UpdatesBatch(s storage.Storage) echo.HandlerFunc {
//	return func(ctx echo.Context) error {
//		var metrics []models.Metrics
//		err := json.NewDecoder(ctx.Request().Body).Decode(&metrics)
//		if err != nil {
//			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
//		}
//		if len(metrics) == 0 {
//			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Batch is empty"))
//		}
//
//		updateGaguges := make([]storage.GaugeMetric, 0, len(metrics))
//		updateCounters := make([]storage.CounterMetric, 0, len(metrics))
//		for _, metric := range metrics {
//			switch metric.MType {
//			case "gauge":
//				if metric.Value == nil {
//					return ctx.String(http.StatusBadRequest, fmt.Sprintf("not value gauge metric"))
//				}
//				//updateGaguges = append(updateGaguges, storage.GaugeMetric{Name: metric.ID, Value: *metric.Value})
//				s.UpdateGauges(updateGaguges)
//			case "counter":
//				if metric.Delta == nil {
//					return ctx.String(http.StatusBadRequest, fmt.Sprintf("not value counter metric"))
//				}
//				//updateCounters = append(updateCounters, storage.CounterMetric{Name: metric.ID, Value: *metric.Delta})
//				s.UpdateCounters(updateCounters)
//			default:
//				return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
//			}
//
//		}
//		//
//		//
//		//err = s.BatchUpdate(metrics)
//		//if err != nil {
//		//	ctx.String(http.StatusInternalServerError, "")
//		//}
//		ctx.Response().Header().Set("Content-Type", "application/json")
//		return ctx.String(http.StatusOK, "")
//	}
//}

func UpdatesBatch(s storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metrics []models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metrics)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}
		if len(metrics) == 0 {
			log.Println("Batch is empty")
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Batch is empty"))
		}
		//
		//updateGaguges := make([]storage.GaugeMetric, 0, len(metrics))
		//updateCounters := make([]storage.CounterMetric, 0, len(metrics))
		//for _, metric := range metrics {
		//	switch metric.MType {
		//	case "gauge":
		//		if metric.Value == nil {
		//			return ctx.String(http.StatusBadRequest, fmt.Sprintf("not value gauge metric"))
		//		}
		//		//updateGaguges = append(updateGaguges, storage.GaugeMetric{Name: metric.ID, Value: *metric.Value})
		//		s.UpdateGauges(updateGaguges)
		//	case "counter":
		//		if metric.Delta == nil {
		//			return ctx.String(http.StatusBadRequest, fmt.Sprintf("not value counter metric"))
		//		}
		//		//updateCounters = append(updateCounters, storage.CounterMetric{Name: metric.ID, Value: *metric.Delta})
		//		s.UpdateCounters(updateCounters)
		//	default:
		//		return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
		//	}
		//
		//}
		//
		//
		err = s.BatchUpdate(metrics)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "")
		}
		encoder := json.NewEncoder(ctx.Response().Writer)
		err = encoder.Encode(metrics[0])
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "error occured on encoding result of batchupdate :%w")
		}
		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.String(http.StatusOK, "")
	}
}

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

func ValueJSON(s storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&metric)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}
		if len(metric.ID) == 0 {
			return c.String(http.StatusNotFound, "")
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
			return c.String(http.StatusBadRequest, "Metric not fount or invalid metric type. Metric type can only be 'gauge' or 'counter'")
		}
		c.Response().Header().Set("Content-Type", "application/json")
		if c.Request().Header.Get("Content-Type") != "application/json" {
			return c.String(http.StatusUnsupportedMediaType, "")
		}
		return c.JSON(http.StatusOK, metric)
	}
}

//func AllMetrics(storage storage.Storage) echo.HandlerFunc {
//	return func(ctx echo.Context) error {
//		ctx.Response().Header().Set("Content-Type", "text/html")
//		m := storage.AllMetrics()
//		result := "Gauge metrics:\n"
//		for name, value := range m.Gauge {
//			result += fmt.Sprintf("- %s = %f\n", name, value)
//		}
//
//		result += "Counter metrics:\n"
//		for name, value := range m.Counter {
//			result += fmt.Sprintf("- %s = %d\n", name, value)
//		}
//		err := ctx.String(http.StatusOK, result)
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
//}

//func GetAllMetrics(storage storage.Storage) echo.HandlerFunc {
//	return func(ctx echo.Context) error {
//		ctx.Response().Header().Set("Content-Type", "text/html")
//		gaugesMetric, err := storage.GetAllGauges()
//		if err != nil {
//			return ctx.String(http.StatusInternalServerError, "")
//		}
//		counterMetric, err := storage.GetAllCounters()
//		if err != nil {
//			return ctx.String(http.StatusInternalServerError, "")
//		}
//
//		fmt.Fprint(ctx.Response().Writer, "<html><body><h1>Metrics</h1><ul>")
//		fmt.Fprint(ctx.Response().Writer, "<h2>Gauges</h2><ul>")
//		for _, metric := range gaugesMetric {
//			fmt.Fprintf(ctx.Response().Writer, "<li>%s: %v</li>", metric.Name, metric.Value)
//		}
//		fmt.Fprint(ctx.Response().Writer, "</ul>")
//
//		fmt.Fprint(ctx.Response().Writer, "</ul><h2>Counters</h2><ul>")
//		for _, metric := range counterMetric {
//			fmt.Fprintf(ctx.Response().Writer, "<li>%s: %v</li>", metric.Name, metric.Value)
//		}
//		fmt.Fprint(ctx.Response().Writer, "</ul></body></html>")
//		//	m := storage.AllMetrics()
//
//		//result := "Gauge metrics:\n"
//		//for _, metric := range gaugesMetric {
//		//	result += fmt.Sprintf("- %s = %f\n", metric.Name, metric.Value)
//		//}
//		//
//		//result += "Counter metrics:\n"
//		//for _, metric := range counterMetric {
//		//	result += fmt.Sprintf("- %s = %d\n", metric.Name, metric.Value)
//		//}
//		err = ctx.String(http.StatusOK, "")
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
//}

func GetAllMetrics(storage storage.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/html")
		gaugesMetric, err := storage.GetAllGauges()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "")
		}
		counterMetric, err := storage.GetAllCounters()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "")
		}
		//	m := storage.AllMetrics()
		var result string

		result += "Gauge metrics:\n"
		for _, metric := range gaugesMetric {
			result += fmt.Sprintf("- %s = %.2f\n", metric.Name, metric.Value)
		}
		result += "Counter metrics:\n"
		for _, metric := range counterMetric {
			result += fmt.Sprintf("- %s = %d\n", metric.Name, metric.Value)
		}
		err = ctx.String(http.StatusOK, result)
		if err != nil {
			return err
		}

		return nil
	}
}

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
			return err
		}

		return nil
	}
}

//func PingDB(db *database.Postgres) echo.HandlerFunc {
//	return func(ctx echo.Context) error {
//		ctx.Response().Header().Set("Content-Type", "text/html")
//		err := database.CheckConnection(db)
//		if err == nil {
//			ctx.String(http.StatusOK, "Connection database is OK")
//		} else {
//			ctx.String(http.StatusInternalServerError, "Connection database is NOT ok")
//		}
//
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
//}
