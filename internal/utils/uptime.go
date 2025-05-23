//go:build !arm.6

package utils

import (
	"fmt"
	"math"
)

func SecondsToClock(seconds int64) string {
	if seconds <= 0 {
		return "00:00:00"
	}
	days := fmt.Sprintf("%d", int(math.Floor(float64(seconds)/86400)))
	hours := fmt.Sprintf("%d", int(math.Floor(math.Mod(float64(seconds), 86400)/3600)))
	mins := fmt.Sprintf("%02d", int(math.Floor(math.Mod(float64(seconds), 3600)/60)))
	secs := fmt.Sprintf("%02d", int(math.Floor(math.Mod(float64(seconds), 60))))
	return days + " days, " + hours + ":" + mins + ":" + secs
}
