package metricsserver

import (
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
)

var _ configuration.Config = &MetricsConfig{}

func NewConfig() *MetricsConfig {
	c := &MetricsConfig{}
	c.Init()
	return c
}

type MetricsConfig struct {
	Enabled bool
	Host    string
	Port    uint16
}

func (mc *MetricsConfig) Init() {
}

func (mc *MetricsConfig) Name() string {
	return "metrics"
}

func (mc *MetricsConfig) Validate() error {
	return nil
}

func (mc *MetricsConfig) SetDefaultValues() error {
	mc.Enabled = false
	return nil
}

func (mc *MetricsConfig) SetExampleValues() error {
	mc.Enabled = false
	mc.Host = "[::]"
	mc.Port = 9100
	return nil
}

func (mc *MetricsConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
