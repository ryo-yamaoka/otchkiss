package result

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentileLatency(t *testing.T) {
	t.Parallel()

	genSeq := func(head, tail int) []float64 {
		var seq []float64
		switch {
		case head < tail:
			for i := head; i <= tail; i++ {
				seq = append(seq, float64(i))
			}
			return seq
		case head > tail:
			for i := head; i >= tail; i-- {
				seq = append(seq, float64(i))
			}
			return seq
		default:
			return []float64{float64(head)}
		}
	}

	testCases := map[string]struct {
		latencies      []float64
		percentile     int
		wantPercentile float64
		wantError      assert.ErrorAssertionFunc
	}{
		"99p sorted": {
			latencies:      genSeq(1, 1000),
			percentile:     99,
			wantPercentile: 990,
			wantError:      assert.NoError,
		},
		"99p no sorted": {
			latencies:      genSeq(1000, 1),
			percentile:     99,
			wantPercentile: 990,
			wantError:      assert.NoError,
		},
		"0p no sorted": {
			latencies:      genSeq(1000, 1),
			percentile:     0,
			wantPercentile: 1,
			wantError:      assert.NoError,
		},
		"100p no sorted": {
			latencies:      genSeq(1000, 1),
			percentile:     100,
			wantPercentile: 1000,
			wantError:      assert.NoError,
		},
		"no result data": {
			latencies:  []float64{},
			percentile: 99,
			wantError:  assert.Error,
		},
		"under limit": {
			latencies:  genSeq(1000, 1),
			percentile: -1,
			wantError:  assert.Error,
		},
		"exceed limit": {
			latencies:  genSeq(1000, 1),
			percentile: 101,
			wantError:  assert.Error,
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			res, err := WithCapacity(len(tc.latencies))
			require.NoError(t, err)
			res.latencies = tc.latencies

			actualPercentile, err := res.PercentileLatency(tc.percentile)
			tc.wantError(t, err)
			assert.Equal(t, tc.wantPercentile, actualPercentile)
		})
	}
}

func TestErrorErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		errs            []error
		wantErrorString string
	}{
		"single": {
			errs:            []error{fmt.Errorf("err1")},
			wantErrorString: "err1",
		},
		"double": {
			errs:            []error{fmt.Errorf("err1"), fmt.Errorf("err2")},
			wantErrorString: "err1, err2",
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			res, err := WithCapacity(len(tc.errs))
			require.NoError(t, err)
			res.errors = tc.errs

			assert.Equal(t, tc.wantErrorString, res.Error())
			diff := cmp.Diff(tc.errs, res.Errors(), cmpopts.EquateErrors())
			assert.Empty(t, diff)
		})
	}
}

func TestRace(t *testing.T) {
	t.Parallel()

	const round = 2048

	var wg sync.WaitGroup
	res, err := WithCapacity(round)
	require.NoError(t, err)
	for i := 0; i < round; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.AppendSuccess(0)
			res.AppendFail(1, fmt.Errorf("err"))
			_ = res.Error()
			_ = res.Errors()
			_ = res.Succeeded()
			_ = res.Failed()
			_ = res.Latencies()
			_, _ = res.PercentileLatency(99)
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(round), res.Succeeded())
	assert.Equal(t, int64(round), res.Failed())
	assert.Len(t, res.Latencies(), round*2)
	assert.Len(t, res.Errors(), round)
}
