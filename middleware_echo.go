package microbot

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/microbot/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func MiddlewareEcho() echo.MiddlewareFunc {
	return MiddlewareEchoWithConfig(DefaultMiddlewareConfig)
}

func MiddlewareEchoWithConfig(config MiddlewareConfig) echo.MiddlewareFunc {
	if config.StackSize == 0 {
		config.StackSize = DefaultMiddlewareConfig.StackSize
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func(begun time.Time) {
				var path string
				if c.Response().Status != http.StatusNotFound {
					path = c.Path()
				}
				if path == "/metrics" {
					return
				}
				s := fmt.Sprintf("%d", c.Response().Status)
				d := time.Since(begun).Nanoseconds() / int64(time.Millisecond)
				ipType := "private"
				if utils.IsPublicIP(c.RealIP()) {
					ipType = "public"
				}

				duration.WithLabelValues(c.Path(), s, c.Request().Method, ipType).Observe(float64(d))
				requests.With(prometheus.Labels{
					"handler": path,
					"status":  s,
					"method":  c.Request().Method,
					"ip_type": ipType,
				}).Inc()
			}(time.Now())

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						fmt.Printf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}
					c.Error(err)
					panics.Inc()
				}
			}()
			return next(c)
		}
	}
}
