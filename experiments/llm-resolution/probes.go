package main

// The probes.
const (
	ProbeBase           = "base"
	ProbeDeletion       = "deletion"
	ProbeControl        = "control"
	ProbeRareShared     = "rare-shared"
	ProbeReconstruction = "reconstruction"
	ProbeRepeat         = "repeat"
)

// ProbeSpec is one probe's browse: its task, and what its tools hide.
type ProbeSpec struct {
	Task     Task
	Redact   map[string]bool
	HideTest string
	Removed  []string
	Note     string
}

// DeletionProbe is not built yet.
func DeletionProbe(t Test, base TestAnswer, answers []string) ProbeSpec { return ProbeSpec{} }

// ControlProbe is not built yet.
func ControlProbe(t Test, base TestAnswer, answers []string, seed int64) ProbeSpec {
	return ProbeSpec{}
}

// RareSharedProbe is not built yet.
func RareSharedProbe(t Test, base TestAnswer, answers []string, reqs []Requirement, b *Baseline) ProbeSpec {
	return ProbeSpec{}
}

// ReconstructionProbe is not built yet.
func ReconstructionProbe(t Test, base TestAnswer) ProbeSpec { return ProbeSpec{} }
