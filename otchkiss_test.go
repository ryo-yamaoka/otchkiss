package otchkiss

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ryo-yamaoka/otchkiss/result"
	"github.com/ryo-yamaoka/otchkiss/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRequesterImpl struct{}

func (tr *testRequesterImpl) Init() error {
	return nil
}

func (tr *testRequesterImpl) RequestOne(_ context.Context) error {
	return nil
}

func (tr *testRequesterImpl) Terminate() error {
	return nil
}

func TestNew(t *testing.T) {
	testCases := map[string]struct {
		requester    Requester
		wantOtchkiss *Otchkiss
		wantError    assert.ErrorAssertionFunc
	}{
		"ok": {
			requester: &testRequesterImpl{},
			wantOtchkiss: &Otchkiss{
				Requester: &testRequesterImpl{},
				Setting: &setting.Setting{
					MaxConcurrent: 1,
					MaxRPS:        1,
					RunDuration:   5 * time.Second,
					WarmUpTime:    5 * time.Second,
				},
			},
			wantError: assert.NoError,
		},
		"ng: nil": {
			requester:    nil,
			wantOtchkiss: nil,
			wantError:    assert.Error,
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			os.Args = []string{"dummy"} // Avoid flag parse error
			ot, err := New(tc.requester)
			tc.wantError(t, err)

			if tc.wantOtchkiss != nil {
				assert.Equal(t, tc.wantOtchkiss.Requester, ot.Requester)
				assert.Equal(t, tc.wantOtchkiss.Setting, ot.Setting)
				assert.NotNil(t, ot.Result)
				return
			}
			assert.Nil(t, ot)
		})
	}
}

func TestFromConfig(t *testing.T) {
	t.Parallel()

	type args struct {
		requester Requester
		setting   *setting.Setting
		cap       int
	}
	testCases := map[string]struct {
		args         args
		wantOtchkiss *Otchkiss
		wantError    assert.ErrorAssertionFunc
	}{
		"ok": {
			args: args{
				requester: &testRequesterImpl{},
				setting: &setting.Setting{
					MaxConcurrent: 2,
					RunDuration:   2 * time.Second,
					WarmUpTime:    10 * time.Second,
				},
				cap: 1,
			},
			wantOtchkiss: &Otchkiss{
				Requester: &testRequesterImpl{},
				Setting: &setting.Setting{
					MaxConcurrent: 2,
					RunDuration:   2 * time.Second,
					WarmUpTime:    10 * time.Second,
				},
			},
			wantError: assert.NoError,
		},
		"ng: requester nil": {
			args: args{
				requester: nil,
				setting: &setting.Setting{
					MaxConcurrent: 2,
					RunDuration:   2 * time.Second,
					WarmUpTime:    10 * time.Second,
				},
				cap: 1,
			},
			wantOtchkiss: nil,
			wantError:    assert.Error,
		},
		"ng: setting nil": {
			args: args{
				requester: &testRequesterImpl{},
				setting:   nil,
				cap:       1,
			},
			wantOtchkiss: nil,
			wantError:    assert.Error,
		},
		"ng: invalid capacity": {
			args: args{
				requester: &testRequesterImpl{},
				setting: &setting.Setting{
					MaxConcurrent: 2,
					RunDuration:   2 * time.Second,
					WarmUpTime:    10 * time.Second,
				},
				cap: -1,
			},
			wantOtchkiss: nil,
			wantError:    assert.Error,
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ot, err := FromConfig(tc.args.requester, tc.args.setting, tc.args.cap)
			tc.wantError(t, err)

			if tc.wantOtchkiss != nil {
				assert.Equal(t, tc.wantOtchkiss.Requester, ot.Requester)
				assert.Equal(t, tc.wantOtchkiss.Setting, ot.Setting)
				require.NotNil(t, ot.Result)
				return
			}
			assert.Nil(t, ot)
		})
	}
}

func TestReport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		setting    *setting.Setting
		templ      string
		wantReport string
		wantError  assert.ErrorAssertionFunc
	}{
		"default": {
			setting: &setting.Setting{
				MaxConcurrent: 1,
				MaxRPS:        1,
				RunDuration:   2 * time.Second,
				WarmUpTime:    3 * time.Second,
			},
			templ:      defaultReportTemplate,
			wantReport: "\n[Setting]\n* warm up time:   3s\n* duration:       2s\n* max concurrent: 1\n* max RPS:        1\n\n[Request]\n* total:      3\n* succeeded:  2\n* failed:     1\n* error rate: 33.3 %\n* RPS:        1.5\n\n[Latency]\n* max: 3,000 ms\n* min: 1,000 ms\n* avg: 2,000 ms\n* med: 1,000 ms\n* 99th percentile: 2,000 ms\n* 90th percentile: 2,000 ms\n",
			wantError:  assert.NoError,
		},
		"user format": {
			setting: &setting.Setting{
				WarmUpTime: 3 * time.Second,
			},
			templ:      "{{.WarmUpTime}}",
			wantReport: "3s",
			wantError:  assert.NoError,
		},
		"empty": {
			setting:    &setting.Setting{},
			templ:      "",
			wantReport: "",
			wantError:  assert.Error,
		},
	}

	for tn, tc := range testCases {
		tn, tc := tn, tc
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r, err := result.WithCapacity(3)
			require.NoError(t, err)
			ot := Otchkiss{
				Result:  r,
				Setting: tc.setting,
			}
			ot.Result.AppendSuccess(1)
			ot.Result.AppendSuccess(2)
			ot.Result.AppendFail(3, errors.New("err1"))

			report, err := ot.TemplateReport(tc.templ)
			tc.wantError(t, err)
			diff := cmp.Diff(tc.wantReport, report)
			assert.Empty(t, diff)

			if tn == "default" {
				report, err := ot.Report()
				tc.wantError(t, err)
				diff := cmp.Diff(tc.wantReport, report)
				assert.Empty(t, diff)
			}
		})
	}
}
