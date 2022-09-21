package setting

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFromDefault(t *testing.T) {
	// DO NOT t.Parallel() because avoid os.Args race condition.

	testCases := map[string]struct {
		args        []string
		wantError   assert.ErrorAssertionFunc
		wantSetting *Setting
	}{
		"default": {
			args:      []string{"test"},
			wantError: assert.NoError,
			wantSetting: &Setting{
				MaxConcurrent: 1,
				RunDuration:   1 * time.Second,
				WarmUpTime:    5 * time.Second,
			},
		},
		"define args": {
			args:      []string{"test", "-p", "2", "-d", "2s", "-w", "2s"},
			wantError: assert.NoError,
			wantSetting: &Setting{
				MaxConcurrent: 2,
				RunDuration:   2 * time.Second,
				WarmUpTime:    2 * time.Second,
			},
		},
		"ng: no unit": {
			args:      []string{"test", "-p", "2", "-d", "2", "-w", "2"},
			wantError: assert.Error,
		},
		"ng: invalid p": {
			args:      []string{"test", "-p", "0", "-d", "2s", "-w", "2s"},
			wantError: assert.Error,
		},
		"ng: invalid d": {
			args:      []string{"test", "-p", "2", "-d", "0s", "-w", "2s"},
			wantError: assert.Error,
		},
		"ng: invalid w": {
			args:      []string{"test", "-p", "2", "-d", "2s", "-w", "-1s"},
			wantError: assert.Error,
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			// DO NOT t.Parallel() because avoid os.Args race condition.

			os.Args = tc.args
			s, err := FromDefaultFlag()
			tc.wantError(t, err)
			diff := cmp.Diff(tc.wantSetting, s)
			assert.Empty(t, diff)
		})
	}
}

func TestNewSetting(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		maxConcurrent int
		runDuration   time.Duration
		warmUpTime    time.Duration
		wantError     assert.ErrorAssertionFunc
		wantSetting   *Setting
	}{
		"ok: default": {
			maxConcurrent: defaultMaxConcurrent,
			runDuration:   defaultRunDuration,
			warmUpTime:    defaultWarmUpTime,
			wantError:     assert.NoError,
			wantSetting: &Setting{
				MaxConcurrent: 1,
				RunDuration:   1 * time.Second,
				WarmUpTime:    5 * time.Second,
			},
		},
		"ok: minimum": {
			maxConcurrent: 1,
			runDuration:   1 * time.Second,
			warmUpTime:    0 * time.Second,
			wantError:     assert.NoError,
			wantSetting: &Setting{
				MaxConcurrent: 1,
				RunDuration:   1 * time.Second,
				WarmUpTime:    0 * time.Second,
			},
		},
		"ng: max concurrent": {
			maxConcurrent: 0,
			runDuration:   defaultRunDuration,
			warmUpTime:    defaultWarmUpTime,
			wantError:     assert.Error,
			wantSetting:   nil,
		},
		"ng: run duration": {
			maxConcurrent: defaultMaxConcurrent,
			runDuration:   0 * time.Second,
			warmUpTime:    defaultWarmUpTime,
			wantError:     assert.Error,
			wantSetting:   nil,
		},
		"ng: warm up time": {
			maxConcurrent: defaultMaxConcurrent,
			runDuration:   defaultRunDuration,
			warmUpTime:    -1 * time.Second,
			wantError:     assert.Error,
			wantSetting:   nil,
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			actual, err := New(tc.maxConcurrent, tc.runDuration, tc.warmUpTime)
			tc.wantError(t, err)
			diff := cmp.Diff(tc.wantSetting, actual)
			assert.Empty(t, diff)
		})
	}
}
