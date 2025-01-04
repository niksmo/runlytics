package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type args struct {
	poll, report time.Duration
	rh           ReportHandler
}

func TestNewCollector(t *testing.T) {

	type test struct {
		name    string
		args    args
		want    *collector
		wantErr error
	}

	fakeHandler := func(data []Metric) {}

	tests := []test{
		{
			name: "New regular collector",
			args: args{
				poll:   time.Duration(1 * time.Second),
				report: time.Duration(2 * time.Second),
				rh:     fakeHandler,
			},
			want: &collector{
				poll:   time.Duration(1 * time.Second),
				report: time.Duration(2 * time.Second),
				rh:     fakeHandler,
			},
			wantErr: nil,
		},
		{
			name: "Poll duration less then 1 sec",
			args: args{
				poll:   time.Duration(500 * time.Millisecond),
				report: time.Duration(10 * time.Second),
				rh:     fakeHandler,
			},
			want:    nil,
			wantErr: ErrMinIntervalValue,
		},
		{
			name: "Report duration less then 1 sec",
			args: args{
				poll:   time.Duration(2 * time.Second),
				report: time.Duration(500 * time.Millisecond),
				rh:     fakeHandler,
			},
			want:    nil,
			wantErr: ErrMinIntervalValue,
		},
		{
			name: "Report and poll duration less then 1 sec",
			args: args{
				poll:   time.Duration(500 * time.Millisecond),
				report: time.Duration(500 * time.Millisecond),
				rh:     fakeHandler,
			},
			want:    nil,
			wantErr: ErrMinIntervalValue,
		},
		{
			name: "Report less then poll duration",
			args: args{
				poll:   time.Duration(2 * time.Second),
				report: time.Duration(1 * time.Second),
				rh:     fakeHandler,
			},
			want:    nil,
			wantErr: ErrReportLessPoll,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collector, err := NewCollector(test.args.poll, test.args.report, test.args.rh)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.IsType(t, test.want.rh, collector.rh)
			assert.Equal(t, test.want.poll, collector.poll)
			assert.Equal(t, test.want.report, collector.report)
			assert.Equal(t, test.want.data, collector.data)
		})
	}
}

func TestCollectorRun(t *testing.T) {
	type want struct {
		reports int
		dataLen int
	}
	type test struct {
		name string
		args args
		want want
	}

	reports := 0

	tests := []test{
		{
			name: "Should handle two times with all metrics",
			args: args{
				poll:   time.Duration(1 * time.Second),
				report: time.Duration(2 * time.Second),
				rh: func(data []Metric) {
					reports++
				},
			},
			want: want{
				reports: 2,
				dataLen: len(memMetrics) + 2,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collector, err := NewCollector(test.args.poll, test.args.report, test.args.rh)
			assert.NoError(t, err)

			go collector.Run()
			time.Sleep(5 * time.Second)

			assert.Equal(t, test.want.reports, reports)
			assert.Equal(t, test.want.dataLen, len(collector.getData()))
		})
	}
}
