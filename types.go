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
	Name          string            `yaml:"name"`
	SaveInterval  time.Duration     `yaml:"saveInterval"`
	StartValue    float64           `yaml:"startValue"` // 策略起始资金，用来计算净值
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
	GetIMR() float64
	GetMMR() float64
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

type Order struct {
	Id           string
	ClientId     string
	Symbol       string
	Price        float64
	Qty          float64
	Side         string
	Status       string
	CancelType   string
	RejectReason string
	AvgPrice     float64
	CumExecQty   float64
	CumExecValue float64
	CumExecFee   float64
	TimeInForce  string
	OrderType    string
	ReduceOnly   bool
	CreatedTime  int64
	UpdatedTime  int64
	Slippage     float64
}

func (order *Order) toInfluxFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["orderId"] = order.Id
	fields["orderClientId"] = order.ClientId
	fields["orderPrice"] = order.Price
	fields["orderQty"] = order.Qty
	fields["orderSide"] = order.Side
	fields["orderStatus"] = order.Status
	fields["orderCancelType"] = order.CancelType
	fields["orderRejectReason"] = order.RejectReason
	fields["orderAvgPrice"] = order.AvgPrice
	fields["orderCumExecQty"] = order.CumExecQty
	fields["orderCumExecValue"] = order.CumExecValue
	fields["orderCumExecFee"] = order.CumExecFee
	fields["orderTimeInForce"] = order.TimeInForce
	fields["orderType"] = order.OrderType
	if order.ReduceOnly {
		fields["orderReduceOnly"] = 1
	} else {
		fields["orderReduceOnly"] = 0
	}
	fields["orderCreatedTime"] = order.CreatedTime
	fields["orderUpdatedTime"] = order.UpdatedTime
	fields["orderSlippage"] = order.Slippage
	return fields
}
