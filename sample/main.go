package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ryo-yamaoka/otchkiss"
	"github.com/ryo-yamaoka/otchkiss/setting"
)

type SampleRequester struct{}

func (tr *SampleRequester) Init() error {
	return nil
}

func (tr *SampleRequester) RequestOne(_ context.Context) error {
	time.Sleep(10 * time.Millisecond) // Substitute for HTTP request
	return nil
}

func (tr *SampleRequester) Terminate() error {
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	st := &setting.Setting{
		RunDuration:   3 * time.Second,
		WarmUpTime:    2 * time.Second,
		MaxConcurrent: 2,
		MaxRPS:        2,
	}
	ot, err := otchkiss.FromConfig(&SampleRequester{}, st, 1_000_000)
	if err != nil {
		return fmt.Errorf("init error: %w", err)
	}

	ctx := context.Background()
	if err := ot.Start(ctx); err != nil {
		return fmt.Errorf("start error: %w", err)
	}
	rep, err := ot.Report()
	if err != nil {
		return fmt.Errorf("report error: %w", err)
	}
	fmt.Println(rep)
	return nil
}
