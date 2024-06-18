package result

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aybabtme/uniplot/histogram"
)

const (
	defaultCapacity = 100_000_000
)

// Result represents to records the results (number of successes or failures and latency) of the tests performed by Otchkiss.
// All implemented methods are thread safe without note.
type Result struct {
	succeeded int64
	failed    int64
	latencies []float64
	sorted    bool
	errors    []error

	latenciesMu sync.Mutex
	errorsMu    sync.Mutex
}

// New returns Result instance by default capacity (100M).
func New() (*Result, error) {
	return new(defaultCapacity)
}

// WithCapacity returns Result instance by given capacity.
// When case of capacity shortage performance may be affected due to automatic memory allocation.
// However too large value may cause OOM.
func WithCapacity(cap int) (*Result, error) {
	return new(cap)
}

func new(cap int) (*Result, error) {
	if cap < 0 {
		return nil, errors.New("capacity must be >= 0")
	}

	return &Result{
		latencies: make([]float64, 0, cap),
		errors:    make([]error, 0, cap),
	}, nil
}

func (r *Result) Error() string {
	r.errorsMu.Lock()
	defer r.errorsMu.Unlock()

	const delimiter = ", "
	var errStr string
	for _, err := range r.errors {
		errStr += err.Error() + delimiter
	}

	return strings.TrimSuffix(errStr, delimiter)
}

func (r *Result) Errors() []error {
	r.errorsMu.Lock()
	defer r.errorsMu.Unlock()
	return r.errors
}

func (r *Result) AppendSuccess(t float64) {
	atomic.AddInt64(&r.succeeded, 1)
	r.appendLatency(t)
}

func (r *Result) AppendFail(t float64, err error) {
	atomic.AddInt64(&r.failed, 1)
	r.appendLatency(t)

	r.errorsMu.Lock()
	defer r.errorsMu.Unlock()
	r.errors = append(r.errors, err)
}

func (r *Result) appendLatency(t float64) {
	r.latenciesMu.Lock()
	defer r.latenciesMu.Unlock()

	r.latencies = append(r.latencies, t)
	r.sorted = false
}

func (r *Result) Succeeded() int64 {
	return atomic.LoadInt64(&r.succeeded)
}

func (r *Result) Failed() int64 {
	return atomic.LoadInt64(&r.failed)
}

func (r *Result) Latencies() []float64 {
	r.latenciesMu.Lock()
	defer r.latenciesMu.Unlock()
	return r.latencies
}

func (r *Result) PercentileLatency(p int) (float64, error) {
	r.latenciesMu.Lock()
	defer r.latenciesMu.Unlock()

	if len(r.latencies) == 0 {
		return 0, errors.New("no result data")
	}

	if p < 0 || p > 100 {
		return 0, errors.New("p must be between 0 and 100")
	}

	if !r.sorted {
		sort.SliceStable(r.latencies, func(i, j int) bool {
			return r.latencies[i] < r.latencies[j]
		})
		r.sorted = true
	}

	switch {
	case p == 0:
		return r.latencies[0], nil
	case p == 100:
		return r.latencies[len(r.latencies)-1], nil
	default:
		idx := float64(len(r.latencies)) * (float64(p) / 100)
		return r.latencies[int(idx-1)], nil
	}
}

func (r *Result) Histogram(bins, width int) (string, error) {
	r.latenciesMu.Lock()
	defer r.latenciesMu.Unlock()

	var buf bytes.Buffer
	hi := histogram.Hist(bins, r.latencies)
	fn := func(v float64) string {
		return time.Duration(v * float64(time.Second)).String()
	}

	if err := histogram.Fprintf(&buf, hi, histogram.Linear(width), fn); err != nil {
		return "", err
	}

	return buf.String(), nil
}
