package bandwidth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

const (
	networkRXBytesPath = "/sys/class/net/%s/statistics/rx_bytes"
	networkTXBytesPath = "/sys/class/net/%s/statistics/tx_bytes"
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
}

func newStatCounter(iface string, db *gorm.DB, statsCallback func(rxMb float64, txMb float64)) *StatCounter {
	return &StatCounter{
		iface: iface,
		db:    db,
		statsCallback: func(rxMb float64, txMb float64) {
			if statsCallback != nil {
				statsCallback(rxMb, txMb)
			}
		},
	}
}

func (s *StatCounter) Start() error {
	if s.running {
		return fmt.Errorf("stat counter already running")
	}
	s.running = true
	var err error
	s.lastRXBytes, err = s.readRXBytes()
	if err != nil {
		fmt.Println("Error reading RX bytes:", err)
		return err
	}
	s.lastTXBytes, err = s.readTXBytes()
	if err != nil {
		fmt.Println("Error reading TX bytes:", err)
		return err
	}
	go func() {
		count := 0
		for s.running {
			time.Sleep(500 * time.Millisecond)
			tunnel, err := models.FindTunnelByInterface(s.db, s.iface)
			if err != nil {
				fmt.Println("Error finding tunnel:", err)
				continue
			}
			if !tunnel.Active {
				return
			}
			rxBytes, err := s.readRXBytes()
			if err != nil {
				fmt.Println("Error reading RX bytes:", err)
				continue
			}
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

			txBytes, err := s.readTXBytes()
			if err != nil {
				fmt.Println("Error reading TX bytes:", err)
				continue
			}
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

			s.statsCallback(float64(s.lastNewRXBytes)/1024/1024, float64(s.lastNewTXBytes)/1024/1024)

			count++
		}
	}()
	return nil
}

func (s *StatCounter) Stop() {
	s.running = false
}

func (s *StatCounter) readRXBytes() (uint64, error) {
	rxBytes, err := os.ReadFile(fmt.Sprintf(networkRXBytesPath, s.iface))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(rxBytes)), 10, 64)
}

func (s *StatCounter) readTXBytes() (uint64, error) {
	txBytes, err := os.ReadFile(fmt.Sprintf(networkTXBytesPath, s.iface))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(txBytes)), 10, 64)
}

type StatCounterManager struct {
	counters         sync.Map
	db               *gorm.DB
	running          bool
	TotalRXMB        float64
	TotalTXMB        float64
	TotalRXBandwidth uint64
	TotalTXBandwidth uint64
}

func NewStatCounterManager(db *gorm.DB) *StatCounterManager {
	return &StatCounterManager{
		db: db,
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
	statCounter := newStatCounter(iface, s.db, s.totalStatsUpdate)
	s.counters.Store(iface, statCounter)
	return statCounter.Start()
}

func (s *StatCounterManager) totalStatsUpdate(rxMb float64, txMb float64) {
	s.TotalRXMB += rxMb
	s.TotalTXMB += txMb
}

func (s *StatCounterManager) updateTotalBandwidth() {
	s.TotalRXBandwidth = 0
	s.TotalTXBandwidth = 0
	for _, counter := range s.GetAll() {
		s.TotalRXBandwidth += counter.RXBandwidth
		s.TotalTXBandwidth += counter.TXBandwidth
	}
}

func (s *StatCounterManager) Remove(iface string) error {
	statCounter, ok := s.counters.Load(iface)
	if !ok {
		return fmt.Errorf("stat counter not found for interface %s", iface)
	}
	statCounter.(*StatCounter).Stop()
	s.counters.Delete(iface)
	return nil
}

func (s *StatCounterManager) Get(iface string) *StatCounter {
	statCounter, ok := s.counters.Load(iface)
	if !ok {
		return nil
	}
	return statCounter.(*StatCounter)
}

func (s *StatCounterManager) GetAll() []*StatCounter {
	var counters []*StatCounter
	s.counters.Range(func(key, value interface{}) bool {
		counters = append(counters, value.(*StatCounter))
		return true
	})
	return counters
}

func (s *StatCounterManager) Stop() {
	s.running = false
	s.counters.Range(func(key, value interface{}) bool {
		value.(*StatCounter).Stop()
		return true
	})
}
