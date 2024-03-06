package xystats

import (
	"github.com/tstrading/xystats/influx/client"
	"sync"
	"time"
)

type InfluxConfig struct {
	Address       string        `yaml:"address"`
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
	Database      string        `yaml:"database"`
	BatchSize     int           `yaml:"batchSize"` // save points in batches if there are multiple records
	WriteInterval time.Duration `yaml:"writeInterval"`
}

type RecordConfig struct {
	StartValue    float64           `yaml:"startValue"` // 策略起始资金，用来计算净值
	Measurement   string            `yaml:"measurement"`
	SymbolMap     map[string]string `yaml:"symbolMap"`
	InfluxConfigs []InfluxConfig    `yaml:"influxConfigs"`
}

type Position interface {
	GetSizeInCoin() float64
	GetEntryPrice() float64
}

type Account interface {
	GetEquity() float64
	GetAvailableBalance() float64
}

type Recorder struct {
	config        RecordConfig
	influxWriters []*InfluxWriter
}

type InfluxWriter struct {
	config       InfluxConfig
	influxClient client.Client
	done         chan interface{}
	points       []*client.Point
	stopped      int64
	allSavedCh   chan interface{}
	allSaved     int64
	mu           sync.Mutex
}
