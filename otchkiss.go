package otchkiss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/ryo-yamaoka/otchkiss/result"
	"github.com/ryo-yamaoka/otchkiss/sema"
	"github.com/ryo-yamaoka/otchkiss/setting"

	humanize "github.com/dustin/go-humanize"
	"go.uber.org/ratelimit"
)

// Requester defines the behavior of the request that Otchkiss performs.
type Requester interface {

	// Init is executed only once before the repeated RequestOne loop.
	// If initialization is unnecessary, do nothing and return nil.
	Init() error

	// RequestOne is executed repeatedly and in parallel.
	// Normally HTTP, gRPC, and other requests are written here.
	// If an error is returned the request is counted as a failure; otherwise, it is counted as a success.
	RequestOne(ctx context.Context) error

	// Terminate is executed only once after the repeated RequestOne loop.
	// If termination is unnecessary, do nothing and return nil.
	Terminate() error
}

type Otchkiss struct {
	Requester Requester
	Setting   *setting.Setting
	Result    *result.Result
}

// New returns Otchkiss instance with default setting.
// By default, the following three command line arguments are parsed and set.
//
//	-p: Specify the number of parallels executions. 0 means unlimited (default: 1, it's not concurrently)
//	-d: Running duration, ex: 300s or 5m etc... (default: 5s)
//	-w: Exclude from results for a given time after startup, ex: 300s or 5m etc... (default: 5s)
//	-r: Specify the max request per second. 0 means unlimited (default: 1)
//
// Note: -p or -r, whichever is smaller blocks the request.
func New(requester Requester) (*Otchkiss, error) {
	s, err := setting.FromDefaultFlag()
	if err != nil {
		return nil, err
	}
	r, err := result.New()
	if err != nil {
		return nil, err
	}

	return new(requester, s, r)
}

// FromConfig returns Otchkiss instance by user specified setting.
// Too large values for resultCapacity may cause OOM (FYI: default value is 100M).
func FromConfig(requester Requester, setting *setting.Setting, resultCapacity int) (*Otchkiss, error) {
	r, err := result.WithCapacity(resultCapacity)
	if err != nil {
		return nil, err
	}
	return new(requester, setting, r)
}

func new(requester Requester, setting *setting.Setting, r *result.Result) (*Otchkiss, error) {
	if requester == nil {
		return nil, errors.New("nil requester")
	}
	if setting == nil {
		return nil, errors.New("nil setting")
	}

	return &Otchkiss{
		Requester: requester,
		Setting:   setting,
		Result:    r,
	}, nil
}

// Start run Otchkiss load testing, and the test follows these steps.
//  1. Run Init()
//  2. Start RequestOne() repeatedly as warm up (it will NOT count as Result)
//  3. Start RequestOne() repeatedly as actual test (it will count as Result)
//  4. End RequestOne() execute and run Terminate()
func (ot *Otchkiss) Start(ctx context.Context) error {
	if err := ot.Requester.Init(); err != nil {
		return fmt.Errorf("failed to initialize requester: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, ot.Setting.RunDuration+ot.Setting.WarmUpTime)
	defer cancel()

	warmUp := make(chan struct{})
	go func() {
		time.Sleep(ot.Setting.WarmUpTime)
		close(warmUp)
	}()

	sem := sema.NewWeighted(int64(ot.Setting.MaxConcurrent))
	rl := ratelimit.NewUnlimited()
	maxRPS := ot.Setting.MaxRPS
	if maxRPS != 0 {
		rl = ratelimit.New(maxRPS, ratelimit.Per(1*time.Second))
	}

	var wg sync.WaitGroup
	for {
		if ctx.Err() != nil {
			break
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}
		rl.Take()

		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			err := ot.Requester.RequestOne(ctx)
			elapsed := time.Since(start) // Do this before error handling to obtain the most accurate time possible.
			sem.Release(1)               // Do this before error handling to release semaphore as soon as possible.

			select {
			case <-warmUp:
				if err != nil {
					ot.Result.AppendFail(elapsed.Seconds(), err)
					return
				}
				ot.Result.AppendSuccess(elapsed.Seconds())
				return
			default:
				return
			}
		}()
	}

	wg.Wait()
	return ot.Requester.Terminate()
}

type ReportParams struct {
	TotalRequests string
	Succeeded     string
	Failed        string
	WarmUpTime    string
	Duration      string
	MaxConcurrent int
	MaxRPS        int
	ErrorRate     string
	RPS           string
	MaxLatency    string
	MinLatency    string
	AvgLatency    string
	MedLatency    string
	Latency99p    string
	Latency90p    string
	Histogram     string
}

// Report outputs result of Otchkiss testing by default template.
func (ot *Otchkiss) Report() (string, error) {
	return ot.report(defaultReportTemplate)
}

// TemplateReport outputs result of Otchkiss testing by user template.
// See `template.go` for a sample.
func (ot *Otchkiss) TemplateReport(template string) (string, error) {
	return ot.report(template)
}

func (ot *Otchkiss) report(templ string) (string, error) {
	if templ == "" {
		return "", errors.New("empty template")
	}

	rp, err := ot.reportParam()
	if err != nil {
		return "", fmt.Errorf("failed to generate report parameters: %w", err)
	}

	tmpl, err := template.New("").Parse(templ)
	if err != nil {
		return "", fmt.Errorf("failed to parse report format: %w", err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, rp); err != nil {
		return "", fmt.Errorf("failed to rendering report: %w", err)
	}

	return buf.String(), nil
}

func (ot *Otchkiss) reportParam() (*ReportParams, error) {
	succeeded := ot.Result.Succeeded()
	failed := ot.Result.Failed()
	total := succeeded + failed

	max, err := ot.Result.PercentileLatency(100)
	if err != nil {
		return nil, fmt.Errorf("failed to get max latency: %w", err)
	}
	min, err := ot.Result.PercentileLatency(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get min latency: %w", err)
	}
	p99, err := ot.Result.PercentileLatency(99)
	if err != nil {
		return nil, fmt.Errorf("failed to get 99p latency: %w", err)
	}
	p90, err := ot.Result.PercentileLatency(90)
	if err != nil {
		return nil, fmt.Errorf("failed to get 90p latency: %w", err)
	}
	p50, err := ot.Result.PercentileLatency(50)
	if err != nil {
		return nil, fmt.Errorf("failed to get 50p latency: %w", err)
	}
	hist, err := ot.Result.Histogram(9, 25)
	if err != nil {
		return nil, fmt.Errorf("failed to generate histogram: %w", err)
	}

	ll := ot.Result.Latencies()
	var avg float64
	for _, l := range ll {
		avg += l
	}
	avg = avg / float64(len(ll))

	return &ReportParams{
		TotalRequests: humanize.Comma(total),
		Succeeded:     humanize.Comma(succeeded),
		Failed:        humanize.Comma(failed),
		WarmUpTime:    ot.Setting.WarmUpTime.String(),
		Duration:      ot.Setting.RunDuration.String(),
		MaxConcurrent: ot.Setting.MaxConcurrent,
		MaxRPS:        ot.Setting.MaxRPS,
		ErrorRate:     humanize.CommafWithDigits(float64(failed)/float64(total)*100, 1),
		RPS:           humanize.CommafWithDigits(float64(total)/ot.Setting.RunDuration.Seconds(), 1),
		MaxLatency:    humanize.CommafWithDigits(max*1000, 1),
		MinLatency:    humanize.CommafWithDigits(min*1000, 1),
		AvgLatency:    humanize.CommafWithDigits(avg*1000, 1),
		MedLatency:    humanize.CommafWithDigits(p50*1000, 1),
		Latency99p:    humanize.CommafWithDigits(p99*1000, 1),
		Latency90p:    humanize.CommafWithDigits(p90*1000, 1),
		Histogram:     hist,
	}, nil
}
