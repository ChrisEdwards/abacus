package ui

// StartupStage enumerates the high-level phases of application initialization.
type StartupStage int

const (
	StartupStageInit StartupStage = iota
	StartupStageVersionCheck
	StartupStageFindingDatabase
	StartupStageLoadingIssues
	StartupStageBuildingGraph
	StartupStageOrganizingTree
	StartupStageReady
)

// StartupReporter receives progress notifications during UI initialization.
// Implementations should be safe for concurrent use.
type StartupReporter interface {
	Stage(stage StartupStage, detail string)
}

// StartupReporterFunc adapts a function to the StartupReporter interface.
type StartupReporterFunc func(stage StartupStage, detail string)

// Stage implements StartupReporter.
func (f StartupReporterFunc) Stage(stage StartupStage, detail string) {
	if f == nil {
		return
	}
	f(stage, detail)
}
