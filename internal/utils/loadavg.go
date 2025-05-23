package utils

import (
	"fmt"
	"os"
	"runtime"
)

type LoadAvg struct {
	OneMin     float64 `json:"one_min"`
	FiveMin    float64 `json:"five_min"`
	FifteenMin float64 `json:"fifteen_min"`
}

func GetLoadAvg() (loadavg LoadAvg, err error) {
	loadavgStr, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}

	_, err = fmt.Sscanf(string(loadavgStr), "%f %f %f", &loadavg.OneMin, &loadavg.FiveMin, &loadavg.FifteenMin)
	if err != nil {
		return
	}

	loadavg.OneMin = loadavg.OneMin / float64(runtime.NumCPU()) * 100
	loadavg.FiveMin = loadavg.FiveMin / float64(runtime.NumCPU()) * 100
	loadavg.FifteenMin = loadavg.FifteenMin / float64(runtime.NumCPU()) * 100

	return
}
