package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hhstu/prometheus-proxy/config"
	"github.com/hhstu/prometheus-proxy/pkg/prom_proxy"
	"net/url"
	"strings"
)

type PrometheusProxy struct{}

func (PrometheusProxy) Proxy(c *gin.Context) {

	promPath := strings.TrimPrefix(c.Request.URL.Path, fmt.Sprintf("/prometheus/"))

	target := config.AppConfig.PrometheusUrl
	uri, _ := url.Parse(target)
	params := getParams(c)

	req := c.Request
	req.URL.Host = uri.Host
	req.Host = uri.Host
	req.URL.Scheme = uri.Scheme
	req.URL.Path = uri.Path + promPath

	prom_proxy.DoRequest(c.Writer, req, params)
}

func getParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for _, key := range config.AppConfig.Params {
		value := c.Query("key1")
		if value != "" {
			params[key] = value
		}
	}
	return params
}
