package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gin-gonic/gin"

)

var (
	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"}, 
	)
)

func InitPrometheus() {
	prometheus.MustRegister(HttpDuration)
}

func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		promHandler := promhttp.Handler()
		promHandler.ServeHTTP(c.Writer, c.Request)
	}
}
