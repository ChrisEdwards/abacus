package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRunWithRuntimeJSONOutput(t *testing.T) {
	runtime := runtimeOptions{
		dbPath:     "/tmp/test.db",
		jsonOutput: true,
	}

	var printerCalled bool
	var gotClient beads.Client
	err := runWithRuntime(runtime, nil, nil, func() startupAnimator {
		t.Fatalf("spinner should not be created in json mode")
		return nil
	}, func(ctx context.Context, client beads.Client) error {
		printerCalled = true
		gotClient = client
		return nil
	}, func(path string) beads.Client {
		if path != runtime.dbPath {
			t.Fatalf("expected db path %q, got %q", runtime.dbPath, path)
		}
		return beads.NewMockClient()
	})
	if err != nil {
		t.Fatalf("runWithRuntime returned error: %v", err)
	}
	if !printerCalled {
		t.Fatal("expected json printer to be called")
	}
	if gotClient == nil {
		t.Fatal("expected client to be passed to printer")
	}
}

func TestRunWithRuntimeSpinnerLifecycle(t *testing.T) {
	spinner := &mockSpinner{}
	runtime := runtimeOptions{
		refreshInterval: time.Second,
	}
	var reporter ui.StartupReporter
	builder := func(cfg ui.Config) (*ui.App, error) {
		reporter = cfg.StartupReporter
		if reporter == nil {
			t.Fatal("expected startup reporter")
		}
		reporter.Stage(ui.StartupStageLoadingIssues, "testing")
		return &ui.App{}, nil
	}

	prog := noopProgram{}
	err := runWithRuntime(runtime, builder, func(app *ui.App) programRunner {
		if !spinner.stopped {
			t.Fatal("expected spinner stopped before program factory")
		}
		return prog
	}, func() startupAnimator {
		return spinner
	}, nil, nil)
	if err != nil {
		t.Fatalf("runWithRuntime returned error: %v", err)
	}
	if len(spinner.stages) == 0 {
		t.Fatal("expected spinner to receive stage updates")
	}
	if !spinner.stopped {
		t.Fatal("expected spinner to stop")
	}
}

func TestRunWithRuntimeStopsSpinnerOnBuilderError(t *testing.T) {
	spinner := &mockSpinner{}
	runtime := runtimeOptions{}
	builder := func(cfg ui.Config) (*ui.App, error) {
		return nil, errors.New("boom")
	}
	err := runWithRuntime(runtime, builder, func(app *ui.App) programRunner {
		t.Fatal("factory should not be called")
		return nil
	}, func() startupAnimator {
		return spinner
	}, nil, nil)
	if err == nil {
		t.Fatal("expected error from builder")
	}
	if spinner.stopCount != 1 {
		t.Fatalf("expected spinner stop count 1, got %d", spinner.stopCount)
	}
}

func TestStartupSpinnerRendersAfterDelay(t *testing.T) {
	var buf bytes.Buffer
	spinner := newCustomStartupSpinner(&buf, 0, time.Millisecond)
	spinner.Stage(ui.StartupStageLoadingIssues, "buffering tests")
	time.Sleep(10 * time.Millisecond)
	spinner.Stop()
	if buf.Len() == 0 {
		t.Fatal("expected spinner output")
	}
}

func TestStartupSpinnerRespectsDelay(t *testing.T) {
	var buf bytes.Buffer
	spinner := newCustomStartupSpinner(&buf, 30*time.Millisecond, time.Millisecond)
	spinner.Stage(ui.StartupStageLoadingIssues, "delayed")
	time.Sleep(10 * time.Millisecond)
	if buf.Len() != 0 {
		t.Fatalf("expected no output before delay, got %q", buf.String())
	}
	time.Sleep(30 * time.Millisecond)
	spinner.Stop()
	if buf.Len() == 0 {
		t.Fatal("expected spinner output after delay")
	}
}

type mockSpinner struct {
	stages    []ui.StartupStage
	stopped   bool
	stopCount int
}

func (m *mockSpinner) Stage(stage ui.StartupStage, detail string) {
	m.stages = append(m.stages, stage)
}

func (m *mockSpinner) Stop() {
	if m.stopped {
		return
	}
	m.stopped = true
	m.stopCount++
}

type noopProgram struct{}

func (noopProgram) Run() (tea.Model, error) {
	return nil, nil
}
