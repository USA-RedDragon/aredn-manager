package bandwidth

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/events"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/vishvananda/netlink"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type StatCounter struct {
	iface          string
	lastRXBytes    uint64
	lastTXBytes    uint64
	lastNewRXBytes uint64
	lastNewTXBytes uint64
	running        bool
	db             *gorm.DB
	RXBandwidth    uint64
	TXBandwidth    uint64
	statsCallback  func(rxMb float64, txMb float64)
	eventsChannel  chan events.Event
}

func newStatCounter(iface string, db *gorm.DB, events chan events.Event, statsCallback func(rxMb float64, txMb float64)) *StatCounter {
	return &StatCounter{
		iface: iface,
		db:    db,
		statsCallback: func(rxMb float64, txMb float64) {
			if statsCallback != nil {
				statsCallback(rxMb, txMb)
			}
		},
		eventsChannel: events,
	}
}

func (s *StatCounter) Start() error {
	if s.running {
		return fmt.Errorf("stat counter already running")
	}
	dev, err := netlink.LinkByName(s.iface)
	if err != nil {
		return fmt.Errorf("error getting link %s by name: %w", s.iface, err)
	}
	s.running = true
	s.lastRXBytes = dev.Attrs().Statistics.RxBytes
	s.lastTXBytes = dev.Attrs().Statistics.TxBytes
	go func() {
		count := 0
		for s.running {
			time.Sleep(500 * time.Millisecond)
			tunnel, err := models.FindTunnelByInterface(s.db, s.iface)
			if err != nil {
				fmt.Println("Error finding tunnel:", err)
				return
			}
			if !tunnel.Active {
				return
			}
			dev, err = netlink.LinkByName(s.iface)
			if err != nil {
				log.Printf("Error getting link %s by name: %v\n", s.iface, err)
				return
			}
			rxBytes := dev.Attrs().Statistics.RxBytes
			if rxBytes < s.lastRXBytes {
				s.lastRXBytes = rxBytes
				continue
			}
			newBytes := rxBytes - s.lastRXBytes
			tunnel.RXBytes += newBytes
			tunnel.TotalRXMB += float64(newBytes) / 1024 / 1024
			if count != 0 && count%2 == 0 {
				tunnel.RXBytesPerSec = s.lastNewRXBytes + newBytes
			}
			s.RXBandwidth = tunnel.RXBytesPerSec
			s.lastNewRXBytes = newBytes
			s.lastRXBytes = rxBytes

			txBytes := dev.Attrs().Statistics.TxBytes
			if txBytes < s.lastTXBytes {
				s.lastTXBytes = txBytes
				continue
			}
			newBytes = txBytes - s.lastTXBytes
			tunnel.TXBytes += newBytes
			tunnel.TotalTXMB += float64(newBytes) / 1024 / 1024
			if count != 0 && count%2 == 0 {
				if count == 100 {
					count = 2
				}
				tunnel.TXBytesPerSec = s.lastNewTXBytes + newBytes
				if err = s.db.Save(tunnel).Error; err != nil {
					fmt.Println("Error saving tunnel:", err)
					continue
				}
			}
			s.TXBandwidth = tunnel.TXBytesPerSec
			s.lastNewTXBytes = newBytes
			s.lastTXBytes = txBytes

			wsTunnel := apimodels.WebsocketTunnelStats{
				ID:               tunnel.ID,
				RXBytesPerSecond: tunnel.RXBytesPerSec,
				TXBytesPerSecond: tunnel.TXBytesPerSec,
				RXBytes:          tunnel.RXBytes,
				TXBytes:          tunnel.TXBytes,
				TotalRXMB:        tunnel.TotalRXMB,
				TotalTXMB:        tunnel.TotalTXMB,
			}

			s.eventsChannel <- events.Event{
				Type: events.EventTypeTunnelStats,
				Data: wsTunnel,
			}

			s.statsCallback(float64(s.lastNewRXBytes)/1024/1024, float64(s.lastNewTXBytes)/1024/1024)

			count++
		}
	}()
	return nil
}

func (s *StatCounter) Stop() {
	s.running = false
}

type StatCounterManager struct {
	counters         sync.Map
	db               *gorm.DB
	running          bool
	TotalRXMB        float64
	TotalTXMB        float64
	TotalRXBandwidth uint64
	TotalTXBandwidth uint64
	eventsChannel    chan events.Event
}

func NewStatCounterManager(db *gorm.DB, events chan events.Event) *StatCounterManager {
	return &StatCounterManager{
		db:            db,
		eventsChannel: events,
	}
}

func (s *StatCounterManager) Start() {
	s.running = true
	go func() {
		time.Sleep(2 * time.Second)
		for s.running {
			s.updateTotalBandwidth()
			time.Sleep(1 * time.Second)
		}
	}()
}

func (s *StatCounterManager) Add(iface string) error {
	statCounter := newStatCounter(iface, s.db, s.eventsChannel, s.totalStatsUpdate)
	_, loaded := s.counters.LoadOrStore(iface, statCounter)
	if loaded {
		return fmt.Errorf("stat counter already exists for interface %s", iface)
	}
	return statCounter.Start()
}

func (s *StatCounterManager) totalStatsUpdate(rxMb float64, txMb float64) {
	s.TotalRXMB += rxMb
	s.TotalTXMB += txMb

	s.eventsChannel <- events.Event{
		Type: events.EventTypeTotalTraffic,
		Data: map[string]float64{
			"RX": s.TotalRXMB,
			"TX": s.TotalTXMB,
		},
	}
}

func (s *StatCounterManager) updateTotalBandwidth() {
	s.TotalRXBandwidth = 0
	s.TotalTXBandwidth = 0
	for _, counter := range s.GetAll() {
		s.TotalRXBandwidth += counter.RXBandwidth
		s.TotalTXBandwidth += counter.TXBandwidth
	}
	s.eventsChannel <- events.Event{
		Type: events.EventTypeTotalBandwidth,
		Data: map[string]uint64{
			"RX": s.TotalRXBandwidth,
			"TX": s.TotalTXBandwidth,
		},
	}
}

func (s *StatCounterManager) Remove(iface string) error {
	statCounter, loaded := s.counters.LoadAndDelete(iface)
	if !loaded {
		return fmt.Errorf("stat counter not found for interface %s", iface)
	}
	sc, ok := statCounter.(*StatCounter)
	if !ok {
		return fmt.Errorf("stat counter type assertion error")
	}
	sc.Stop()
	return nil
}

func (s *StatCounterManager) Get(iface string) *StatCounter {
	statCounter, ok := s.counters.Load(iface)
	if !ok {
		return nil
	}
	sc, ok := statCounter.(*StatCounter)
	if !ok {
		return nil
	}
	return sc
}

func (s *StatCounterManager) GetAll() []*StatCounter {
	var counters []*StatCounter
	s.counters.Range(func(_, value interface{}) bool {
		sc, ok := value.(*StatCounter)
		if !ok {
			return true
		}
		counters = append(counters, sc)
		return true
	})
	return counters
}

func (s *StatCounterManager) Stop() error {
	s.running = false
	errGrp := errgroup.Group{}
	s.counters.Range(func(_, value interface{}) bool {
		sc, ok := value.(*StatCounter)
		if !ok {
			return true
		}
		errGrp.Go(func() error {
			sc.Stop()
			return nil
		})
		return true
	})

	return errGrp.Wait()
}
