package metrics

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//nolint:gochecknoglobals
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
	if config.Metrics.Enabled {
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
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              fmt.Sprintf(":%d", config.Metrics.Port),
			ReadHeaderTimeout: 5 * time.Second,
		}
		err := server.ListenAndServe()
		if err != nil {
			slog.Error("Error starting metrics server", "error", err)
		}
	}
}
