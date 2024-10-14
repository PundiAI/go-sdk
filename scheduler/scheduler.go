package scheduler

import (
	"context"
	"math"
	"time"

	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/tool"
)

const (
	maxSleepMs = 60 * 1000
)

type taskFunc func(ctx context.Context) error

type Scheduler interface {
	Run(ctx context.Context) error
}

type scheduler struct {
	logger   log.Logger
	config   Config
	taskFunc taskFunc
}

func New(logger log.Logger, config Config, taskFunc taskFunc) Scheduler {
	return &scheduler{
		logger:   logger.With("scheduler", config.Name),
		config:   config,
		taskFunc: taskFunc,
	}
}

func (t *scheduler) Run(ctx context.Context) error {
	if !t.config.Enabled {
		return nil
	}
	t.logger.Infof("start %s service", t.config.Name)
	errCount := uint16(0)
	normalSleepTime := t.config.Interval
	sleepTime := normalSleepTime
	timer := time.NewTimer(normalSleepTime)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			t.logger.Debugf("execute %s service", t.config.Name)
			err := t.taskFunc(ctx)

			errCount, sleepTime = calcSleepTime(normalSleepTime, errCount, err)
			if err != nil {
				t.logger.Warnf("execute %s error: %v, errorCount: [%d], sleep: [%s]", t.config.Name, err, errCount, sleepTime.String())
			}
			if errCount >= t.config.MaxErrCount {
				return err
			}

			timer.Reset(sleepTime)
		case <-ctx.Done():
			t.logger.Infof("stop %s service", t.config.Name)
			return nil
		}
	}
}

func calcSleepTime(normalSleep time.Duration, errCount uint16, err error) (uint16, time.Duration) {
	if err == nil {
		return 0, normalSleep
	}
	if !tool.IsIgnoreException(err) {
		errCount++
	}
	sleepMs := int64(float64(normalSleep.Milliseconds()) * math.Pow(1.4, float64(errCount)))
	if sleepMs > maxSleepMs {
		sleepMs = maxSleepMs
	}
	sleepTime := time.Duration(sleepMs) * time.Millisecond
	return errCount, sleepTime
}
