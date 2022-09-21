package setting

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromDefault(t *testing.T) {
	t.Parallel()

	s, err := FromDefaultFlag()
	require.NoError(t, err)
	assert.Equal(t, 1, s.MaxConcurrent)
	assert.Equal(t, 1*time.Second, s.RunDuration)
	assert.Equal(t, 5*time.Second, s.WarmUpTime)
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
