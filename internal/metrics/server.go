package metrics

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func CreateMetricsServer(config *config.Config) {
	port := config.MetricsPort
	if port != 0 {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}
}
