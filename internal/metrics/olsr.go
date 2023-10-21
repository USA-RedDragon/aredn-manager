package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OLSRLinkAsymmetryTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_asymmetry_time",
		Help: "OLSR Link Asymmetry Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkHelloTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_hello_time",
		Help: "OLSR Link Hello Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkHysteresis = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_hysteresis",
		Help: "OLSR Link Hysteresis",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLastHelloTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_last_hello_time",
		Help: "OLSR Link Last Hello Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLinkCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_link_cost",
		Help: "OLSR Link Link Cost",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLinkQuality = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_link_quality",
		Help: "OLSR Link Link Quality",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLossHelloInterval = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_loss_hello_interval",
		Help: "OLSR Link Loss Hello Interval",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLossMultiplier = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_loss_multiplier",
		Help: "OLSR Link Loss Multiplier",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLossTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_loss_time",
		Help: "OLSR Link Loss Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkLostLinkTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_lost_link_time",
		Help: "OLSR Link Lost Link Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkNeighborLinkQuality = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_neighbor_link_quality",
		Help: "OLSR Link Neighbor Link Quality",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkPending = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_pending",
		Help: "OLSR Link Pending",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkSeqno = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_seqno",
		Help: "OLSR Link Seqno",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkSeqnoValid = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_seqno_valid",
		Help: "OLSR Link Seqno Valid",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkSymmetryTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_symmetry_time",
		Help: "OLSR Link Symmetry Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkValidityTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_validity_time",
		Help: "OLSR Link Validity Time",
	}, []string{"device", "local_ip", "remote_ip"})
	OLSRLinkVTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olsr_link_vtime",
		Help: "OLSR Link VTime",
	}, []string{"device", "local_ip", "remote_ip"})
)

func OLSRWatcher() {
	for {
		resp, err := http.DefaultClient.Get("http://localhost:9090/links")
		if err != nil {
			fmt.Printf("OLSRWatcher: Unable to get links: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		var links apimodels.OlsrdLinks
		err = json.NewDecoder(resp.Body).Decode(&links)
		if err != nil {
			fmt.Printf("OLSRWatcher: Unable to decode links: %v\n", err)
			time.Sleep(1 * time.Second)
			resp.Body.Close()
			continue
		}

		for _, link := range links.Links {
			pending := 0
			if link.Pending {
				pending = 1
			}

			seqnoValid := 0
			if link.SeqnoValid {
				seqnoValid = 1
			}

			OLSRLinkAsymmetryTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.AsymmetryTime))
			OLSRLinkHelloTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.HelloTime))
			OLSRLinkHysteresis.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.Hysteresis))
			OLSRLinkLastHelloTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LastHelloTime))
			OLSRLinkLinkCost.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LinkCost))
			OLSRLinkLinkQuality.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LinkQuality))
			OLSRLinkLossHelloInterval.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LossHelloInterval))
			OLSRLinkLossMultiplier.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LossMultiplier))
			OLSRLinkLossTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LossTime))
			OLSRLinkLostLinkTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.LostLinkTime))
			OLSRLinkNeighborLinkQuality.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.NeighborLinkQuality))
			OLSRLinkPending.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(pending))
			OLSRLinkSeqno.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.Seqno))
			OLSRLinkSeqnoValid.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(seqnoValid))
			OLSRLinkSymmetryTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.SymmetryTime))
			OLSRLinkValidityTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.ValidityTime))
			OLSRLinkVTime.WithLabelValues(link.OLSRInterface, link.LocalIP, link.RemoteIP).Set(float64(link.VTime))
		}

		resp.Body.Close()
		time.Sleep(1 * time.Second)
	}
}
