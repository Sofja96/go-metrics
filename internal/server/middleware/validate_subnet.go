package middleware

import (
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// ValidateTrustedSubnet - middleware для проверки доверенной подсети по заголовку X-Real-IP.
func ValidateTrustedSubnet(trustedSubnet string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if trustedSubnet == "" {
				log.Info("No trusted subnet configured, skipping validation.")
				return next(c)
			}

			realIP := c.Request().Header.Get("X-Real-IP")
			if realIP == "" {
				log.Warn("Missing X-Real-IP header")
				return c.String(http.StatusBadRequest, "Missing X-Real-IP header")
			}

			_, cidr, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				log.Error("Invalid trusted subnet configuration", "error", err)
				return c.String(http.StatusBadRequest, "Invalid trusted subnet configuration")
			}

			ip := net.ParseIP(realIP)
			if !cidr.Contains(ip) {
				log.Warn("Access denied: IP not in trusted subnet",
					"ip ", ip,
					" trustedSubnet ", trustedSubnet)
				return c.String(http.StatusForbidden, "Access denied: IP not in trusted subnet")
			}

			log.Info("Access granted: IP in trusted subnet",
				"ip ", ip.String(),
				" trustedSubnet ", trustedSubnet)

			if err = next(c); err != nil {
				c.Error(err)
			}

			return err
		}

	}
}
