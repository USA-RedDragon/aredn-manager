package config_test

import (
	"testing"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
)

func TestNoop(t *testing.T) {
	t.Parallel()
	conf := config.Config{}
	conf.Debug = true
	if conf.Debug != true {
		t.Error("Debug should be true")
	}
	t.Log(conf.ToString())
}
