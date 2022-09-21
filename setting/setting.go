package setting

import (
	"errors"
	"flag"
	"os"
	"time"
)

const (
	defaultMaxConcurrent = 1
	defaultRunDuration   = 1 * time.Second
	defaultWarmUpTime    = 5 * time.Second
)

type Setting struct {
	// MaxConcurrent defines how many RequestOne should concurrently running.
	MaxConcurrent int

	// RunDuration defines how long RequestOne should continue to run.
	// During the time specified here, the request results are included in the Result.
	RunDuration time.Duration

	// WarmUpTime defines how long RequestOne not included in the measurement should continue to run.
	// During the time specified here, the request results are NOT included in the Result.
	WarmUpTime time.Duration
}

// New returns Setting instance made by user defined config.
func New(maxConcurrent int, runDuration, warmUpTime time.Duration) (*Setting, error) {
	return newSetting(maxConcurrent, runDuration, warmUpTime)
}

// FromDefaultConfig returns Setting by flag or default value config.
func FromDefaultFlag() (*Setting, error) {
	c := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	maxConcurrent := c.Int("p", defaultMaxConcurrent, "Specify the number of parallels executions (default: 1, it's not concurrently)")
	runDuration := c.Duration("d", defaultRunDuration, "Running duration, ex: 300s or 5m etc... (default: 1s)")
	warmUpTime := c.Duration("w", defaultWarmUpTime, "Exclude from results for a given time after startup, ex: 300s or 5m etc... (default: 5s)")
	if err := c.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return newSetting(*maxConcurrent, *runDuration, *warmUpTime)
}

func newSetting(maxConcurrent int, runDuration, warmUpTime time.Duration) (*Setting, error) {
	if !(maxConcurrent > 0) {
		return nil, errors.New("max concurrent must be > 0")
	}
	if !(runDuration > 0*time.Second) {
		return nil, errors.New("run duration must be > 0 sec")
	}
	if !(warmUpTime >= 0*time.Second) {
		return nil, errors.New("warm up time must be >= 0 sec")
	}

	return &Setting{
		MaxConcurrent: maxConcurrent,
		RunDuration:   runDuration,
		WarmUpTime:    warmUpTime,
	}, nil
}
