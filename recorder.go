package xystats

import (
	"context"
	"errors"
	"fmt"
	"github.com/tstrading/xystats/influx/client"
	"github.com/tstrading/xystats/logger"
	"math"
	"time"
)

func (r *Recorder) Save(
	timestamp time.Time,
	xAccount, yAccount Account,
	xPositionMap, yPositionMap map[string]Position,
	xMidPriceMap map[string]float64,
	yMidPriceMap map[string]float64,
	xVolume24HMap map[string]float64, // 监控短期异常交易量
	yVolume24HMap map[string]float64,
	xVolume30DMap map[string]float64,
	yVolume30DMap map[string]float64,
	xMaxLeverage, yMaxLeverage float64, // 风控指标
	maxOpenValueMap map[string]float64, // 风控指标
) error {

	if xAccount == nil {
		return errors.New("xAccount is nil")
	}

	xTotal24HVolume := 0.0
	yTotal24HVolume := 0.0
	xTotal30DVolume := 0.0
	yTotal30DVolume := 0.0
	xTotalOpenValue := 0.0
	yTotalOpenValue := 0.0
	xTotalLongValue := 0.0
	xTotalShortValue := 0.0
	yTotalLongValue := 0.0
	yTotalShortValue := 0.0
	totalUnhedgedValue := 0.0

	for xSymbol, ySymbol := range r.config.SymbolMap {
		xPosition, ok1 := xPositionMap[xSymbol]
		yPosition, ok2 := yPositionMap[ySymbol]
		xMidPrice, ok3 := xMidPriceMap[xSymbol]
		yMidPrice, ok4 := yMidPriceMap[ySymbol]
		xVolume24H, ok5 := xVolume24HMap[xSymbol]
		yVolume24H, ok6 := yVolume24HMap[xSymbol]
		xVolume30D, ok7 := xVolume30DMap[xSymbol]
		yVolume30D, ok8 := yVolume30DMap[xSymbol]
		maxOpenValue, ok9 := maxOpenValueMap[xSymbol]
		if !ok1 || xPosition == nil {
			return fmt.Errorf("xPosition for %s is missing", xSymbol)
		}
		if !ok2 || yPosition == nil {
			return fmt.Errorf("yPosition for %s is missing", ySymbol)
		}
		if !ok3 || xMidPrice == 0 {
			return fmt.Errorf("xMidPrice for %s is missing", xSymbol)
		}
		if !ok4 || yMidPrice == 0 {
			return fmt.Errorf("yMidPrice for %s is missing", ySymbol)
		}
		if !ok5 {
			return fmt.Errorf("xVolume24H for %s is missing", xSymbol)
		}
		if !ok6 {
			return fmt.Errorf("yVolume24H for %s is missing", ySymbol)
		}
		if !ok7 {
			return fmt.Errorf("xVolume30D for %s is missing", xSymbol)
		}
		if !ok8 {
			return fmt.Errorf("yVolume30D for %s is missing", ySymbol)
		}
		if !ok9 {
			return fmt.Errorf("maxOpenValue for %s is missing", xSymbol)
		}

		xSize := xPosition.GetSizeInCoin()
		ySize := yPosition.GetSizeInCoin()
		xPrice := xPosition.GetEntryPrice()
		yPrice := yPosition.GetEntryPrice()
		if xPrice == 0 {
			xPrice = xMidPrice
		}
		if yPrice == 0 {
			yPrice = yMidPrice
		}
		xValue := xSize * xPrice
		yValue := ySize * yPrice
		unhedgedValue := (xSize + ySize) * (xMidPrice + yMidPrice) / 2
		xTotalOpenValue += math.Abs(xValue)
		yTotalOpenValue += math.Abs(yValue)
		totalUnhedgedValue += math.Abs(unhedgedValue)
		if xValue >= 0 {
			xTotalLongValue += xValue
		} else {
			xTotalShortValue -= xValue
		}
		if yValue >= 0 {
			yTotalLongValue += yValue
		} else {
			yTotalShortValue -= yValue
		}
		xTotal24HVolume += xVolume24H
		yTotal24HVolume += yVolume24H
		xTotal30DVolume += xVolume30D
		yTotal30DVolume += yVolume30D
		fields := make(map[string]interface{})
		fields["xSize"] = xSize
		fields["ySize"] = ySize
		fields["xValue"] = xValue
		fields["yValue"] = yValue
		fields["maxOpenValue"] = maxOpenValue
		fields["xPrice"] = xPrice
		fields["yPrice"] = yPrice
		fields["xMidPrice"] = xMidPrice
		fields["yMidPrice"] = yMidPrice
		fields["unhedgedValue"] = unhedgedValue
		fields["xVolume24H"] = xVolume24H
		fields["yVolume24H"] = yVolume24H
		fields["xVolume30D"] = xVolume30D
		fields["yVolume30D"] = yVolume30D
		pt, err := client.NewPoint(
			r.config.Name,
			map[string]string{
				"xSymbol": xSymbol,
				"ySymbol": ySymbol,
				"type":    "symbol",
			},
			fields,
			timestamp,
		)
		if err != nil {
			logger.Warn("client.NewPoint for %s, %s error %v", xSymbol, ySymbol, err)
		} else {
			for _, iw := range r.influxWriters {
				iw.PushPoint(pt)
			}
		}
	}

	xBalance := xAccount.GetEquity()
	xAvailableBalance := xAccount.GetAvailableBalance()
	yBalance := xBalance
	yAvailableBalance := xAvailableBalance
	xIMR := xAccount.GetIMR()
	xMMR := xAccount.GetMMR()
	yIMR := xIMR
	yMMR := xMMR
	totalBalance := xBalance
	if yAccount != nil {
		yBalance = yAccount.GetEquity()
		yAvailableBalance = yAccount.GetAvailableBalance()
		yIMR = yAccount.GetIMR()
		yMMR = yAccount.GetMMR()
		totalBalance = xBalance + yBalance
	}

	fields := make(map[string]interface{})

	fields["xIMR"] = xIMR
	fields["yIMR"] = yIMR
	fields["xMMR"] = xMMR
	fields["yMMR"] = yMMR
	fields["totalUnhedgedValue"] = totalUnhedgedValue
	fields["totalBalance"] = totalBalance
	fields["yBalance"] = yBalance
	fields["xBalance"] = xBalance
	fields["yAvailable"] = xAvailableBalance
	fields["xAvailable"] = yAvailableBalance
	fields["xTotal30DVolume"] = xTotal30DVolume
	fields["yTotal30DVolume"] = yTotal30DVolume
	fields["xTotal24HVolume"] = xTotal24HVolume
	fields["yTotal24HVolume"] = yTotal24HVolume
	fields["xMaxLeverage"] = xMaxLeverage
	fields["yMaxLeverage"] = yMaxLeverage
	fields["xTotalOpenValue"] = xTotalOpenValue
	fields["yTotalOpenValue"] = yTotalOpenValue
	if xBalance > 0 {
		fields["xLeverage"] = xTotalOpenValue / xBalance
		fields["xTurnover"] = xTotal24HVolume / xBalance
		if xTotalOpenValue > 0 {
			fields["xTotalLongValue"] = xTotalLongValue
			fields["xTotalShortValue"] = xTotalShortValue
			fields["xImbalance"] = (xTotalLongValue - xTotalShortValue) / xTotalOpenValue
		}
	}
	if yBalance > 0 {
		fields["yLeverage"] = yTotalOpenValue / yBalance
		fields["yTurnover"] = yTotal24HVolume / yBalance
		if yTotalOpenValue > 0 {
			fields["yTotalLongValue"] = yTotalLongValue
			fields["yTotalShortValue"] = yTotalShortValue
			fields["yImbalance"] = (yTotalLongValue - yTotalShortValue) / yTotalOpenValue
		}
	}
	if totalBalance > 0 {
		fields["xyLeverage"] = (xTotalOpenValue + yTotalOpenValue) / totalBalance
		fields["xyTurnover"] = (xTotal24HVolume + yTotal24HVolume) / totalBalance
	}
	fields["startValue"] = r.config.StartValue
	if r.config.StartValue > 0 {
		fields["netWorth"] = totalBalance / r.config.StartValue
	}
	pt, err := client.NewPoint(
		r.config.Name,
		map[string]string{
			"type": "summary",
		},
		fields,
		time.Now().UTC(),
	)
	if err != nil {
		logger.Debugf("client.NewPoint error %v", err)
	} else {
		for _, iw := range r.influxWriters {
			iw.PushPoint(pt)
		}
	}
	return nil
}

func NewRecorder(ctx context.Context, config RecordConfig) (*Recorder, error) {
	r := &Recorder{config: config}
	for _, ic := range config.InfluxConfigs {
		iw, err := NewInfluxWriter(ctx, ic)
		if err != nil {
			return nil, err
		}
		r.influxWriters = append(r.influxWriters, iw)
	}
	return r, nil
}
