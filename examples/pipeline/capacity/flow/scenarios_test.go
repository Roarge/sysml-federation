package flow

import (
	"testing"

	"github.com/Roarge/sysml-federation/internal/scenario"
)

// TestScenarios runs the Gherkin scenarios under features/, one subtest per
// acceptance criterion (AD-0034).
func TestScenarios(t *testing.T) { scenario.Run(t, steps) }
