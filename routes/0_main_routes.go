package routes

import (
	"fmt"
	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/hhstu/prometheus-proxy/apis"
	"github.com/hhstu/prometheus-proxy/config"
	"github.com/hhstu/prometheus-proxy/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
)

func init() {
	gin.SetMode(config.AppConfig.Webserver.Mode)
}

func Routes() *gin.Engine {
	r := gin.Default()

	// 基础监控
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.GET("/metrics", prometheusHandler())
	r.GET("/debug/pprof/*pprof", gin.WrapH(http.DefaultServeMux))
	r.Use(ginprom.PromMiddleware(nil))
	r.Use(HandlerRecover)
	prom := PrometheusProxy{}
	promGroup := r.Group("/prometheus/api/v1/")
	promGroup.Any("query", prom.Proxy)
	promGroup.Any("query_range", prom.Proxy)
	promGroup.Any("alerts", prom.Proxy)
	promGroup.Any("alertmanagers", prom.Proxy)
	promGroup.Any("labels", prom.Proxy)
	promGroup.Any("label/:name/values", prom.Proxy)
	promGroup.Any("series", prom.Proxy)
	promGroup.Any("targets", prom.Proxy)
	promGroup.Any("targets/metadata", prom.Proxy)
	promGroup.Any("metadata", prom.Proxy)
	promGroup.Any("rules", prom.Proxy)
	return r
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func HandlerRecover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("panic: internal error %+v", r)
			debug.PrintStack()
			c.AbortWithStatusJSON(http.StatusInternalServerError, apis.Response{
				Status: http.StatusInternalServerError,
				Msg:    fmt.Sprintf("系统内部错误"),
			})
		}
	}()
	c.Next()
}
