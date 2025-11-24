package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"abacus/internal/ui"
)

const defaultSpinnerInterval = 120 * time.Millisecond

type spinnerEvent struct {
	stage  ui.StartupStage
	detail string
}

type startupSpinner struct {
	writer        io.Writer
	delay         time.Duration
	frameInterval time.Duration
	frames        []rune

	events chan spinnerEvent
	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once

	mu       sync.Mutex
	frameIdx int
}

func newStartupSpinner(w io.Writer, delay time.Duration) *startupSpinner {
	return newCustomStartupSpinner(w, delay, defaultSpinnerInterval)
}

func newCustomStartupSpinner(w io.Writer, delay, frameInterval time.Duration) *startupSpinner {
	if w == nil {
		w = io.Discard
	}
	sp := &startupSpinner{
		writer:        w,
		delay:         delay,
		frameInterval: frameInterval,
		frames:        []rune{'|', '/', '-', '\\'},
		events:        make(chan spinnerEvent, 8),
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
	go sp.loop()
	return sp
}

func (s *startupSpinner) Stage(stage ui.StartupStage, detail string) {
	if s == nil {
		return
	}
	select {
	case <-s.stopCh:
		return
	default:
	}
	select {
	case s.events <- spinnerEvent{stage: stage, detail: detail}:
	default:
	}
}

func (s *startupSpinner) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		close(s.stopCh)
		<-s.doneCh
	})
}

func (s *startupSpinner) loop() {
	defer close(s.doneCh)

	var delayCh <-chan time.Time
	if s.delay > 0 {
		timer := time.NewTimer(s.delay)
		defer timer.Stop()
		delayCh = timer.C
	} else {
		delayCh = nil
	}

	ticker := time.NewTicker(s.frameInterval)
	defer ticker.Stop()

	var current spinnerEvent
	hasStage := false
	visible := s.delay == 0

	for {
		select {
		case <-s.stopCh:
			if visible {
				s.clearLine()
			}
			return
		case ev := <-s.events:
			current = ev
			hasStage = true
			if visible {
				s.render(current)
			}
		case <-ticker.C:
			if visible && hasStage {
				s.render(current)
			}
		case <-delayCh:
			delayCh = nil
			if hasStage {
				visible = true
				s.render(current)
			}
		}
	}
}

func (s *startupSpinner) render(ev spinnerEvent) {
	frame := s.nextFrame()
	message := formatStageMessage(ev.stage, ev.detail)
	_, _ = fmt.Fprintf(s.writer, "\r\033[2K%c %s", frame, message)
}

func (s *startupSpinner) clearLine() {
	_, _ = fmt.Fprint(s.writer, "\r\033[2K")
}

func (s *startupSpinner) nextFrame() rune {
	s.mu.Lock()
	defer s.mu.Unlock()
	frame := s.frames[s.frameIdx%len(s.frames)]
	s.frameIdx++
	return frame
}

var wittyStageMessages = map[ui.StartupStage]string{
	ui.StartupStageFindingDatabase: "Counting beads...",
	ui.StartupStageLoadingIssues:   "Sliding issues into place...",
	ui.StartupStageBuildingGraph:   "Stringing dependencies...",
	ui.StartupStageOrganizingTree:  "Aligning the rods...",
	ui.StartupStageReady:           "Polishing the abacus...",
}

func formatStageMessage(stage ui.StartupStage, detail string) string {
	witty := wittyStageMessages[stage]
	if strings.TrimSpace(witty) == "" {
		witty = "Calculating priority positions..."
	}
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return witty
	}
	return fmt.Sprintf("%s - %s", witty, detail)
}
