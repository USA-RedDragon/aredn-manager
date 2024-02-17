package metrics

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	AREDNMeshRF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "node_details_meshrf",
		Help: "AREDN Mesh RF Enabled",
	})
	AREDNInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_aredn_info",
		Help: "AREDN Node Info",
	}, []string{
		"board_id",
		"description",
		"firmware_version",
		"gridsquare",
		"lat",
		"lon",
		"model",
		"node",
		"tactical",
	})
)

func CreateMetricsServer(config *config.Config, version string) {
	// We don't use RF, so we set it to 0
	AREDNMeshRF.Set(0)
	AREDNInfo.WithLabelValues(
		"0x0000",
		"AREDN Cloud Tunnel",
		version,
		config.Gridsquare,
		config.Latitude,
		config.Longitude,
		"Virtual",
		config.ServerName,
		"",
	).Set(1)
	port := config.MetricsPort
	if port != 0 {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}
}
