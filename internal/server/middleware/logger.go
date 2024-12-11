package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func WithLogging(sugar zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			herr := next(c)
			if herr != nil {
				c.Error(herr)
			}

			resp := c.Response()
			req := c.Request()
			duration := time.Since(start)
			// отправляем сведения о запросе в zap
			sugar.Infow(
				"Request",
				"method", req.Method,
				"uri", req.RequestURI,
				"duration", duration.String(),
				"status", resp.Status,
				"size", resp.Size,
			)
			return nil
		}
	}
}
