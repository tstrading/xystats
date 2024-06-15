package xystats

import (
	"context"
	"github.com/tstrading/xystats/influx/client"
	"github.com/tstrading/xystats/logger"
	"sync/atomic"
	"time"
)

func (iw *InfluxWriter) Done() chan interface{} {
	return iw.done
}

func (iw *InfluxWriter) Stop() {
	if atomic.CompareAndSwapInt64(&iw.stopped, 0, 1) {
		close(iw.done)
		<-iw.allSavedCh
	}
}

func (iw *InfluxWriter) PushPoint(pt *client.Point) {
	defer iw.pushMu.Unlock()
	iw.pushMu.Lock()
	iw.pushedPoints = append(iw.pushedPoints, pt)
}

func (iw *InfluxWriter) PushPoints(pts []*client.Point) {
	defer iw.pushMu.Unlock()
	iw.pushMu.Lock()
	iw.pushedPoints = append(iw.pushedPoints, pts...)
}

func (iw *InfluxWriter) save(saveAll bool) {
	iw.saveMu.Lock()
	iw.pushMu.Lock()
	if len(iw.pushedPoints) > 0 {
		iw.savingPoints = append(iw.savingPoints, iw.pushedPoints...)
		iw.pushedPoints = iw.pushedPoints[:0]
	}
	iw.pushMu.Unlock()
	iw.saveMu.Unlock()
	if saveAll {
		defer func() {
			if atomic.CompareAndSwapInt64(&iw.allSaved, 0, 1) {
				close(iw.allSavedCh)
			}
		}()
	}
	retryCount := 0
	maxRetryCount := 10
	for {
		if retryCount > maxRetryCount {
			logger.Warnf("%p save failed, retryCount %d > maxRetryCount %d", iw, retryCount, maxRetryCount)
			return
		}
		saveEnd := -1
		iw.saveMu.Lock()
		if len(iw.savingPoints) > 0 {
			if len(iw.savingPoints) > iw.config.BatchSize {
				saveEnd = iw.config.BatchSize
			} else {
				saveEnd = len(iw.savingPoints)
			}
		}
		if saveEnd == -1 {
			iw.saveMu.Unlock()
			return
		}
		savingPoints := iw.savingPoints[:saveEnd]
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  iw.config.Database,
			Precision: "ns",
		})
		if err != nil {
			logger.Warnf("%p client.NewBatchPoints error %v", iw, err)
			retryCount++
			iw.saveMu.Unlock()
			time.Sleep(time.Second * 3)
			continue
		}

		bp.AddPoints(savingPoints)
		err = iw.influxClient.Write(bp)
		if err != nil {
			logger.Warnf("%p iw.influxClient.Write error %v", iw, err)
			retryCount++
			iw.saveMu.Unlock()
			time.Sleep(time.Second * 3)
			continue
		}

		iw.savingPoints = iw.savingPoints[saveEnd:]
		iw.saveMu.Unlock()

		if !saveAll {
			return
		}
	}
}

func (iw *InfluxWriter) watchPoints(ctx context.Context) {
	saveTimer := time.NewTimer(iw.config.WriteInterval)

	defer func() {
		iw.save(true)
		saveTimer.Stop()
		iw.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-iw.done:
			return
		case <-saveTimer.C:
			iw.save(false)
			saveTimer.Reset(iw.config.WriteInterval)
			break
		}
	}
}

func NewInfluxWriter(ctx context.Context, config InfluxConfig) (*InfluxWriter, error) {
	if config.BatchSize <= 100 {
		config.BatchSize = 100
	}
	if config.WriteInterval <= time.Second*15 {
		config.WriteInterval = time.Second * 15
	}
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		Timeout:  time.Minute * 5,
	})
	if err != nil {
		return nil, err
	}
	iw := &InfluxWriter{
		config:       config,
		influxClient: influxClient,
		done:         make(chan interface{}, 1),
		pushedPoints: make([]*client.Point, 0),
		savingPoints: make([]*client.Point, 0),
		stopped:      0,
		allSaved:     0,
		allSavedCh:   make(chan interface{}),
	}
	go iw.watchPoints(ctx)
	return iw, nil
}
