package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pundiai/go-sdk/log"
)

func Test_scheduler_Run(t1 *testing.T) {
	type fields struct {
		logger log.Logger
		cfg    Config
		exec   func(ctx context.Context) error
	}
	tests := []struct {
		name    string
		fields  fields
		newCtx  func() (context.Context, context.CancelFunc)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			fields: fields{
				logger: log.NewNopLogger(),
				cfg: Config{
					Enabled:     true,
					Name:        "test",
					Interval:    100 * time.Millisecond,
					MaxErrCount: 1,
				},
				exec: func(ctx context.Context) error {
					return nil
				},
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 200*time.Millisecond)
			},
			wantErr: assert.NoError,
		},
		{
			name: "error",
			fields: fields{
				logger: log.NewNopLogger(),
				cfg: Config{
					Enabled:     true,
					Name:        "test",
					Interval:    50 * time.Millisecond,
					MaxErrCount: 2,
				},
				exec: func(ctx context.Context) error {
					return assert.AnError
				},
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 200*time.Millisecond)
			},
			wantErr: assert.Error,
		},
		{
			name: "context cancel",
			fields: fields{
				logger: log.NewNopLogger(),
				cfg: Config{
					Enabled:     true,
					Name:        "test",
					Interval:    10 * time.Millisecond,
					MaxErrCount: 1,
				},
				exec: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			ctx, cancel := tt.newCtx()
			defer func() {
				if cancel != nil {
					cancel()
				}
			}()
			tt.wantErr(t1, New(tt.fields.logger, tt.fields.cfg, tt.fields.exec).Run(ctx))
		})
	}
}

func TestCalcSleepTime(t *testing.T) {
	testCases := []struct {
		name            string
		normalSleep     time.Duration
		errCount        uint16
		err             error
		expectSleepTime time.Duration
		expectErrCount  uint16
	}{
		{
			name:            "err is nil, errCount is 0",
			normalSleep:     100 * time.Millisecond,
			errCount:        0,
			err:             nil,
			expectSleepTime: 100 * time.Millisecond,
			expectErrCount:  0,
		},
		{
			name:            "err is nil, errCount is 1",
			normalSleep:     100 * time.Millisecond,
			errCount:        1,
			err:             nil,
			expectSleepTime: 100 * time.Millisecond,
			expectErrCount:  0,
		},
		{
			name:            "err is not nil, errCount is 0, first error",
			normalSleep:     100 * time.Millisecond,
			errCount:        0,
			err:             errors.New("test"),
			expectSleepTime: 140 * time.Millisecond,
			expectErrCount:  1,
		},
		{
			name:            "err is not nil, errCount is 8, test continue exponential backoff",
			normalSleep:     100 * time.Millisecond,
			errCount:        8,
			err:             errors.New("test"),
			expectSleepTime: 2066 * time.Millisecond,
			expectErrCount:  9,
		},
		{
			name:            "err is not nil, errCount is 10, test maxSleepMs",
			normalSleep:     10 * time.Second,
			errCount:        10,
			err:             errors.New("test"),
			expectSleepTime: maxSleepMs * time.Millisecond,
			expectErrCount:  11,
		},
	}

	for i, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errCount, sleepTime := calcSleepTime(testCase.normalSleep, testCase.errCount, testCase.err)
			assert.Equal(t, testCase.expectErrCount, errCount, "test case %d", i)
			assert.Equal(t, testCase.expectSleepTime, sleepTime, "test case %d", i)
		})
	}
}
